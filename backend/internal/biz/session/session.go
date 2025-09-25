package httpapi

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"stream-server/internal/biz/session/store_pool"
)

// ChunkManager инкапсулирует работу с чанками: загрузка, поиск позиции, переходы вперёд
type ChunkManager struct {
	store    *store_pool.ChunkStore
	streamID uuid.UUID
	meta     store_pool.StreamMeta

	chunk     *store_pool.Chunk // текущий чанк (держим refs)
	pos       int               // позиция внутри текущего чанка
	seq       int64             // следующая желаемая sequence (двигается строго вперёд)
	emptyRuns int               // подряд "пустых" попаданий по чанкам (для страховки)
}

func NewChunkManager(store *store_pool.ChunkStore, streamID uuid.UUID, meta store_pool.StreamMeta) *ChunkManager {
	return &ChunkManager{store: store, streamID: streamID, meta: meta, seq: meta.MinSeq}
}

// get гарантирует кадр с sequence >= cm.seq (если существует)
func (cm *ChunkManager) get(ctx context.Context) (bool, store_pool.Frame) {
	if cm.seq > cm.meta.MaxSeq {
		return false, store_pool.Frame{}
	}
	// Если чанка нет/исчерпан/устарел, то взять чанк, содержащий cm.seq (или ближайший следующий)
	if cm.chunk == nil || cm.seq > cm.chunk.Frames[len(cm.chunk.Frames)-1].Seq {
		if cm.chunk != nil {
			cm.store.ReleaseChunk(cm.chunk)
			cm.chunk = nil
		}
		chunk, err := cm.store.GetChunk(ctx, cm.streamID, cm.meta.MinSeq, cm.seq)
		if err != nil {
			return false, store_pool.Frame{}
		}
		p := sort.Search(len(chunk.Frames), func(i int) bool {
			return chunk.Frames[i].Seq >= cm.seq
		})
		if p >= len(chunk.Frames) {
			// Если этом чанке нет нужной sequence, то перескочим к следующему чанку и повторим
			cm.seq = chunk.StartSeq + cm.store.ChunkSize()
			cm.store.ReleaseChunk(chunk)
			cm.chunk = nil
			cm.pos = 0
			cm.emptyRuns++
			return cm.get(ctx)
		}
		cm.chunk = chunk
		cm.pos = p
		cm.emptyRuns = 0
	}
	return true, cm.chunk.Frames[cm.pos]
}

// advance — сдвинуть курсор на следующий кадр, обновив seq
func (cm *ChunkManager) advance() {
	if cm.chunk == nil {
		return
	}
	cm.seq = cm.chunk.Frames[cm.pos].Seq + 1
	cm.pos++
}

// release — отпустить текущий чанк (refs--)
func (cm *ChunkManager) release() {
	if cm.chunk != nil {
		cm.store.ReleaseChunk(cm.chunk)
		cm.chunk = nil
	}
}

// StreamSession временная шкала + отправка в ws
type StreamSession struct {
	ctx       context.Context
	conn      *websocket.Conn
	store     *store_pool.ChunkStore
	meta      store_pool.StreamMeta
	cm        *ChunkManager
	base      time.Time     // старт времени воспроизведения
	interval  time.Duration // интервал между кадрами (например, 40ms)
	slots     int64         // пройдено слотов по времени (скипы + отправки)
	delivered int64         // реально отправлено кадров
}

func NewStreamSession(ctx context.Context, conn *websocket.Conn, store *store_pool.ChunkStore, meta store_pool.StreamMeta, streamID uuid.UUID) *StreamSession {
	return &StreamSession{
		ctx:   ctx,
		conn:  conn,
		store: store,
		meta:  meta,
		cm:    NewChunkManager(store, streamID, meta),
		base:  time.Now(),
		// в задании указано воспроизводить кадры с частотой 25fps
		// но также можно использовать значение стрима, если использовать строку ниже
		//interval:  time.Duration(meta.IntervalMS) * time.Millisecond,
		interval:  40 * time.Millisecond, // 25fps
		slots:     0,
		delivered: 0,
	}
}

// Run — главный цикл: догоняем временную шкалу скипами, затем в текущем слоте отправляем один кадр
// Завершаемся по концу данных (seq > max_seq) или по ошибке/разрыву соединения
// Внимание: мы НЕ требуем "delivered == Count". Это сознательно, так как важно отсутствие запаздывания стрима
func (s *StreamSession) Run() error {
	defer s.cm.release()

	const emptyChunkGuard = 3 // страховка от редких "вакуумов" в конце

	for {
		// конец данных
		if s.cm.seq > s.meta.MaxSeq {
			return nil
		}

		elapsed := time.Since(s.base)              // сколько слотов времени уже прошло на текущий момент
		targetSlots := int64(elapsed / s.interval) // сколько "должно было" быть кадров
		// защита от глюка
		if targetSlots < 0 {
			targetSlots = 0
		}

		// Догоняем временную шкалу скипами (без отправки)
		for s.slots < targetSlots && s.cm.seq <= s.meta.MaxSeq {
			ok, _ := s.cm.get(s.ctx)
			if !ok {
				if s.cm.emptyRuns >= emptyChunkGuard {
					return nil // хвостовые дырки, тогда завершаемся
				}
				// нет доступных кадров в этом шаге, поэтому просто попробуем на следующей итерации
				break
			}
			s.cm.advance()
			s.slots++ // слот времени пропускаем
		}

		// Текущий слот — пытаемся отправить один кадр (если он есть)
		if s.cm.seq > s.meta.MaxSeq {
			return nil
		}
		ok, f := s.cm.get(s.ctx)
		if ok {
			_ = s.conn.SetWriteDeadline(time.Now().Add(2 * time.Second)) // защита от медленных клиентов
			if err := s.conn.WriteMessage(websocket.BinaryMessage, f.Data); err != nil {
				return err // клиент ушёл/таймаут
			}
			s.cm.advance()
			s.delivered++
		} else {
			// Нечего отправлять в этот слот, такое возможно при больших дырках
			if s.cm.emptyRuns >= emptyChunkGuard {
				return nil
			}
		}
		s.slots++ // слот времени завершён (либо скип, либо отправка)

		// Доспать до начала следующего слота (прерываемый контекстом)
		nextSlotTime := s.base.Add(time.Duration(s.slots) * s.interval)
		if d := time.Until(nextSlotTime); d > 0 {
			timer := time.NewTimer(d)
			select {
			case <-s.ctx.Done():
				timer.Stop()
				return s.ctx.Err()
			case <-timer.C:
			}
		}
	}
}
