package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	session_pool "stream-server/internal/biz/session"
	"stream-server/internal/biz/session/store_pool"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// В продакшене можно ограничить CheckOrigin по доменам фронта/хедеру
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// readerPump — читает входящие кадры, чтобы обрабатывать ping/pong/close и поддерживать read-deadline
// Мы ничего не ждём от клиента, но без этого gorilla НЕ вызовет PongHandler
func readerPump(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	// Защита от злоупотребления: максимум 64К на входящее сообщение (нам ничего не шлют)
	conn.SetReadLimit(64 << 10)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		// продлеваем read-deadline при каждом Pong
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return // клиент отвалился/закрылся
		}
	}
}

func WSStreamHandler(store *store_pool.ChunkStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Валидация id
		idStr, err := extractID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		streamID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "bad stream id", http.StatusBadRequest)
			return
		}

		// Метаданные (min/max/count/interval) — фиксируем "снимок" стрима на момент запроса
		ctx := r.Context()
		meta, err := store.LoadStreamMeta(ctx, streamID)
		if err != nil {
			http.Error(w, "stream not found", http.StatusNotFound)
			return
		}
		if meta.Count == 0 || meta.MaxSeq < meta.MinSeq {
			// 204 если кадров нет — до апгрейда.
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Апгрейд до WS
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		conn.EnableWriteCompression(false) // JPEG уже сжат, компрессия лишь нагружает CPU

		// Запускаем reader и ждём его завершения через канал
		readerDone := make(chan struct{})
		go readerPump(conn, readerDone)

		// Запуск сессии
		session := session_pool.NewStreamSession(ctx, conn, store, meta, streamID)
		runErr := session.Run()

		if runErr == nil {
			// Нормально закрываем поток (конец данных)
			_ = conn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "end of stream"),
				time.Now().Add(1*time.Second))
			return
		}

		// Закрываем соединение (если уже закрыто — ок), это добьёт readerPump
		_ = conn.Close()

		// Немного подождём выхода readerPump (чтобы не оставлять горутину висеть)
		select {
		case <-readerDone:
		case <-time.After(100 * time.Millisecond):
			// не критично, просто перестраховка
		}
		// Если ошибка отправки/разрыв, то просто выходим (дефер закроет сокет)
	}
}

func extractID(r *http.Request) (string, error) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		return "", fmt.Errorf("bad path: %s", r.URL.Path)
	}
	return parts[3], nil
}
