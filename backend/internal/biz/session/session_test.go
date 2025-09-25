package httpapi

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"stream-server/internal/biz/session/store_pool"
)

func newStore(limit int64, chunkN int64) *store_pool.ChunkStore {
	sizes := []int{32 << 10, 64 << 10}
	return store_pool.NewChunkStore(nil, sizes, limit, chunkN)
}

func TestChunkManagerSequentialAcrossChunks(t *testing.T) {
	cs := newStore(1<<20, 2) // small chunks: 2 frames per chunk
	stream := uuid.New()
	meta := store_pool.StreamMeta{
		ID:         stream,
		IntervalMS: 40,
		MinSeq:     10,
		MaxSeq:     13,
	}

	// Prepare two chunks: idx0 -> seq 10,11 ; idx1 -> seq 12,13
	ch0 := &store_pool.Chunk{StartSeq: 10, Frames: []store_pool.Frame{
		{Seq: 10, Data: make([]byte, 1)},
		{Seq: 11, Data: make([]byte, 1)},
	}}
	for _, f := range ch0.Frames {
		ch0.BytesLen += int64(len(f.Data))
		ch0.BytesCap += int64(cap(f.Data))
	}
	store_pool.InjectChunkForTest(store_pool.ChunkKey{Stream: stream, Index: 0}, ch0, cs)

	ch1 := &store_pool.Chunk{StartSeq: 12, Frames: []store_pool.Frame{
		{Seq: 12, Data: make([]byte, 1)},
		{Seq: 13, Data: make([]byte, 1)},
	}}
	for _, f := range ch1.Frames {
		ch1.BytesLen += int64(len(f.Data))
		ch1.BytesCap += int64(cap(f.Data))
	}
	store_pool.InjectChunkForTest(store_pool.ChunkKey{Stream: stream, Index: 1}, ch1, cs)

	cm := NewChunkManager(cs, stream, meta)
	ctx := context.Background()

	expected := []int64{10, 11, 12, 13}
	var got []int64
	for i := 0; i < 4; i++ {
		ok, f := cm.get(ctx)
		if !ok {
			t.Fatalf("expected frame at step %d", i)
		}
		got = append(got, f.Seq)
		cm.advance()
	}
	cm.release()

	if len(got) != len(expected) {
		t.Fatalf("got %v want %v", got, expected)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("got %v want %v", got, expected)
		}
	}
}
