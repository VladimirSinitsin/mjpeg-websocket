package store_pool

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/singleflight"
)

// -------- Конфиг для store --------

// MaxFrameBytes Максимально допустимый размер одного JPEG. Всё больше — считаем "аномалией" и кадр пропускаем
// В проде это должно быть конфигурируемо, здесь зададим здравое значение по умолчанию
const MaxFrameBytes = 4 << 20 // 4 MiB

// PressureGuardFactor Когда после эвикта cap-бюджет всё ещё больше лимита в N раз — не грузим новые чанки
const PressureGuardFactor = 2 // 200% лимита — отказываем в загрузке чанка

// LoadChunkTimeout Таймаут на загрузку чанка из БД (для защиты от повисшей БД)
const LoadChunkTimeout = 500 * time.Millisecond

// Sizes Бакеты для пула
var Sizes = []int{32 << 10, 64 << 10, 128 << 10, 256 << 10, 512 << 10, 1 << 20, 2 << 20, 4 << 20}

// StreamMeta метаданные стрима
type StreamMeta struct {
	ID         uuid.UUID
	IntervalMS int32
	MinSeq     int64
	MaxSeq     int64
	Count      int64
}

// Frame — один JPEG-кадр. Data — буфер из ByteBucketPool (len — реальный размер, cap — размер ведра)
type Frame struct {
	Seq  int64
	Data []byte
	Mime string
}

// Chunk — порция кадров, загружаемая одним SQL и разделяемая многими клиентами
type Chunk struct {
	StartSeq int64
	Frames   []Frame
	BytesLen int64 // сумма len(Data) — для метрик
	BytesCap int64 // сумма cap(Data) — честный объём RAM

	// atomics
	refs    int32  // сколько клиентов держат чанк
	evicted uint32 // снят из LRU (ожидает освобождение при refs==0)
	freed   uint32 // буферы уже возвращены в пулы
}

type ChunkKey struct {
	Stream uuid.UUID
	Index  int64 // floor((seq - minSeq)/chunkN)
}

type lruEntry struct {
	key   ChunkKey
	chunk *Chunk
}

// ChunkStore — LRU-кэш чанков с бюджетом по cap-байтам и одинарной загрузкой (singleflight)
// Все изменения LRU и учёта памяти под мьютексом
type ChunkStore struct {
	db    *pgxpool.Pool
	group singleflight.Group

	mu    sync.Mutex
	lru   *list.List
	items map[ChunkKey]*list.Element

	usedLenB int64 // метрика (сумма len)
	usedCapB int64 // реальный бюджет (сумма cap)
	limitB   int64 // лимит RAM по cap (например, 512<<20)

	chunkN int64           // кадров в чанке (например, 256)
	pool   *ByteBucketPool // пул буферов

	frameSlicePool sync.Pool // пул []Frame
}

func NewChunkStore(db *pgxpool.Pool, sizes []int, limitCapBytes int64, chunkFrames int64) *ChunkStore {
	return &ChunkStore{
		db:     db,
		lru:    list.New(),
		items:  make(map[ChunkKey]*list.Element),
		limitB: limitCapBytes,
		chunkN: chunkFrames,
		pool:   NewByteBucketPool(sizes),
		frameSlicePool: sync.Pool{
			New: func() any { return make([]Frame, 0, int(chunkFrames)) },
		},
	}
}

func (cs *ChunkStore) ChunkSize() int64 {
	return cs.chunkN
}

func (cs *ChunkStore) getFrameSlice() []Frame {
	fs := cs.frameSlicePool.Get().([]Frame)
	if cap(fs) < int(cs.chunkN) {
		return make([]Frame, 0, int(cs.chunkN))
	}
	return fs[:0]
}

func (cs *ChunkStore) putFrameSlice(fs []Frame) {
	cs.frameSlicePool.Put(fs[:0]) // len = 0, сохраняем cap
}

// loadChunk — копируем payload из pgx в наш буфер (из пула) — копия обязательна (pgx реюзит буфер)
// Таймаут на чтение из БД (LoadChunkTimeout) нужен, чтобы подвисшая БД не вешала сессию насмерть
// Слишком большие кадры (>MaxFrameBytes) пропускаем (логически)
func (cs *ChunkStore) loadChunk(ctx context.Context, stream uuid.UUID, startSeq int64) (*Chunk, error) {
	dbCtx, cancel := context.WithTimeout(ctx, LoadChunkTimeout)
	defer cancel()

	rows, err := cs.db.Query(dbCtx, `
        SELECT sequence, payload, mime_type
        FROM frames
        WHERE stream_id = $1 AND sequence >= $2
        ORDER BY sequence
        LIMIT $3
    `, stream, startSeq, cs.chunkN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	frames := cs.getFrameSlice()
	var totalLen, totalCap int64

	for rows.Next() {
		var seq int64
		var src []byte
		var mime string
		if err = rows.Scan(&seq, &src, &mime); err != nil {
			// откат уже взятых буферов, возврат ошибки, так как иначе получим битый чанк
			for i := range frames {
				cs.pool.Put(frames[i].Data)
			}
			cs.putFrameSlice(frames)
			return nil, err
		}
		// Пропускаем аномально большие кадры (защита RAM/сети).
		if len(src) > MaxFrameBytes {
			continue
		}
		dst := cs.pool.Get(len(src)) // len == реальный размер, cap == ведро
		copy(dst, src)
		frames = append(frames, Frame{Seq: seq, Data: dst[:len(src)], Mime: mime})
		totalLen += int64(len(src))
		totalCap += int64(cap(dst)) // лимитируем по cap (bucket), честно к RSS
	}
	if err = rows.Err(); err != nil {
		for i := range frames {
			cs.pool.Put(frames[i].Data)
		}
		cs.putFrameSlice(frames)
		return nil, err
	}

	return &Chunk{
		StartSeq: startSeq,
		Frames:   frames,
		BytesLen: totalLen,
		BytesCap: totalCap,
	}, nil
}

// GetChunk — вернуть чанк по желаемой sequence; увеличивает refs — вызывающий обязан ReleaseChunk
// Против переполнения RAM: если после эвикта usedCapB > limit*PressureGuardFactor, то возвращаем ошибку
func (cs *ChunkStore) GetChunk(ctx context.Context, stream uuid.UUID, minSeq, wantSeq int64) (*Chunk, error) {
	if wantSeq < minSeq {
		wantSeq = minSeq
	}
	idx := (wantSeq - minSeq) / cs.chunkN
	key := ChunkKey{Stream: stream, Index: idx}

	// 1) Попытка взять из LRU за O(1)
	cs.mu.Lock()
	if el := cs.items[key]; el != nil {
		cs.lru.MoveToFront(el)
		chunk := el.Value.(*lruEntry).chunk
		atomic.AddInt32(&chunk.refs, 1)
		cs.mu.Unlock()
		return chunk, nil
	}
	cs.mu.Unlock()

	// 2) Загрузка (dedup через singleflight) + LRU put
	v, err, _ := cs.group.Do(fmt.Sprintf("%s:%d", stream, idx), func() (any, error) {
		// double-check под замком
		cs.mu.Lock()
		if el := cs.items[key]; el != nil {
			cs.lru.MoveToFront(el)
			chunk := el.Value.(*lruEntry).chunk
			cs.mu.Unlock()
			return chunk, nil
		}
		cs.mu.Unlock()

		// Мягкая защита: если уже в 2+ раза выше бюджета — откажем до освобождения.
		cs.mu.Lock()
		if cs.usedCapB > cs.limitB*PressureGuardFactor {
			cs.mu.Unlock()
			return nil, errors.New("cache pressure: cap budget exceeded")
		}
		cs.mu.Unlock()

		startSeq := minSeq + idx*cs.chunkN
		chunk, err := cs.loadChunk(ctx, stream, startSeq)
		if err != nil {
			return nil, err
		}

		cs.mu.Lock()
		el := cs.lru.PushFront(&lruEntry{key: key, chunk: chunk})
		cs.items[key] = el
		cs.usedLenB += chunk.BytesLen
		cs.usedCapB += chunk.BytesCap

		// Эвиктим по cap, насколько возможно
		cs.evictLocked()

		// Если и после эвикта бюджет всё ещё дико выше лимита, то не пойдём дальше
		if cs.usedCapB > cs.limitB*PressureGuardFactor {
			// Уберём только что вставленный элемент обратно
			delete(cs.items, key)
			cs.lru.Remove(el)
			cs.usedLenB -= chunk.BytesLen
			cs.usedCapB -= chunk.BytesCap
			cs.mu.Unlock()

			// Вернём буферы (чтобы не протечь)
			for i := range chunk.Frames {
				cs.pool.Put(chunk.Frames[i].Data)
				chunk.Frames[i].Data = nil
			}
			cs.putFrameSlice(chunk.Frames)
			return nil, errors.New("cache pressure: over budget after eviction")
		}

		cs.mu.Unlock()
		return chunk, nil
	})
	if err != nil {
		return nil, err
	}

	chunk := v.(*Chunk)
	atomic.AddInt32(&chunk.refs, 1)
	return chunk, nil
}

// ReleaseChunk — уменьшаем refs; если чанк эвикнут и refs==0, то освобождаем буферы в пулы
func (cs *ChunkStore) ReleaseChunk(chunk *Chunk) {
	if chunk == nil {
		return
	}
	if atomic.AddInt32(&chunk.refs, -1) == 0 && atomic.LoadUint32(&chunk.evicted) == 1 {
		cs.mu.Lock()
		cs.tryFinalizeChunkLocked(chunk)
		cs.mu.Unlock()
	}
}

// evictLocked — снимаем хвостовые элементы LRU, пока usedCapB > limitB
// Буферы реально освобождаются только при refs==0 (иначе ждём ReleaseChunk)
func (cs *ChunkStore) evictLocked() {
	for cs.usedCapB > cs.limitB {
		el := cs.lru.Back()
		if el == nil {
			break
		}
		entry := el.Value.(*lruEntry)
		chunk := entry.chunk

		delete(cs.items, entry.key)
		cs.lru.Remove(el)
		atomic.StoreUint32(&chunk.evicted, 1)

		// Пробуем освободить прямо сейчас, если никто не держит
		cs.tryFinalizeChunkLocked(chunk)
	}
}

// tryFinalizeChunkLocked — вернуть буферы в пулы и скорректировать учёт (под мьютексом)
func (cs *ChunkStore) tryFinalizeChunkLocked(chunk *Chunk) {
	if atomic.LoadUint32(&chunk.freed) == 1 {
		return
	}
	// освобождаем только если снят с LRU и никто не держит
	if atomic.LoadUint32(&chunk.evicted) != 1 || atomic.LoadInt32(&chunk.refs) != 0 {
		return
	}
	for i := range chunk.Frames {
		if chunk.Frames[i].Data != nil {
			cs.pool.Put(chunk.Frames[i].Data)
			chunk.Frames[i].Data = nil
		}
	}
	cs.usedLenB -= chunk.BytesLen
	cs.usedCapB -= chunk.BytesCap
	if cs.usedLenB < 0 {
		cs.usedLenB = 0
	}
	if cs.usedCapB < 0 {
		cs.usedCapB = 0
	}
	cs.putFrameSlice(chunk.Frames)
	chunk.Frames = nil
	atomic.StoreUint32(&chunk.freed, 1)
}

// LoadStreamMeta — корректный LEFT JOIN + GROUP BY с MIN/MAX/COUNT.
// Для VOD-семантики мы фиксируем "снимок" на момент подключения (это ок).
func (cs *ChunkStore) LoadStreamMeta(ctx context.Context, id uuid.UUID) (StreamMeta, error) {
	var m StreamMeta
	m.ID = id

	row := cs.db.QueryRow(ctx, `
        SELECT
            s.frame_interval_ms,
            COALESCE(MIN(f.sequence), 0)  AS min_seq,
            COALESCE(MAX(f.sequence), -1) AS max_seq,
            COALESCE(COUNT(f.sequence), 0) AS cnt
        FROM streams s
        LEFT JOIN frames f ON f.stream_id = s.id
        WHERE s.id = $1
        GROUP BY s.id, s.frame_interval_ms
    `, id)
	if err := row.Scan(&m.IntervalMS, &m.MinSeq, &m.MaxSeq, &m.Count); err != nil {
		return m, errors.New("stream not found")
	}
	return m, nil
}
