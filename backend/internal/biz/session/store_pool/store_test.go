package store_pool

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

// helper to create a store with given limit and chunk size
func newTestStore(limit int64, chunkN int64) *ChunkStore {
	sizes := []int{32 << 10, 64 << 10, 128 << 10}
	return NewChunkStore(nil, sizes, limit, chunkN)
}

// helper to allocate a frame with given size and seq using the store's pool
func makeFrame(cs *ChunkStore, seq int64, size int) Frame {
	b := cs.pool.Get(size)
	for i := range b {
		b[i] = 1
	}
	return Frame{
		Seq:  seq,
		Data: b[:size],
		Mime: "image/jpeg",
	}
}

// helper to push a chunk directly into LRU/map (unit test within same package)
func addChunk(cs *ChunkStore, key ChunkKey, frames []Frame) *Chunk {
	ch := &Chunk{
		StartSeq: key.Index*cs.chunkN + 0,
		Frames:   frames,
	}
	for _, f := range frames {
		ch.BytesLen += int64(len(f.Data))
		ch.BytesCap += int64(cap(f.Data))
	}
	cs.mu.Lock()
	el := &lruEntry{
		key:   key,
		chunk: ch,
	}

	le := cs.lru.PushFront(el)
	cs.items[key] = le
	cs.usedLenB += ch.BytesLen
	cs.usedCapB += ch.BytesCap
	cs.mu.Unlock()
	return ch
}

func TestGetChunk_LRUHitAndEvictFinalize(t *testing.T) {
	cs := newTestStore(1<<20, 4)
	stream := uuid.New()
	minSeq := int64(100)
	// idx = (want-min)/chunkN = (101-100)/4 = 0
	key := ChunkKey{Stream: stream, Index: 0}
	frames := []Frame{
		makeFrame(cs, 100, 10),
		makeFrame(cs, 101, 20),
		makeFrame(cs, 102, 30),
	}
	ch := addChunk(cs, key, frames)

	// LRU hit path
	got, err := cs.GetChunk(context.Background(), stream, minSeq, 101)
	if err != nil {
		t.Fatalf("GetChunk error: %v", err)
	}
	if got != ch {
		t.Fatalf("expected same chunk pointer on LRU hit")
	}
	if refs := atomic.LoadInt32(&got.refs); refs <= 0 {
		t.Fatalf("refs not incremented")
	}

	// Release reference
	cs.ReleaseChunk(got)

	// Force eviction by setting tiny limit and calling evictLocked
	cs.mu.Lock()
	cs.limitB = 0
	cs.evictLocked()
	cs.mu.Unlock()

	// After eviction (refs==0) the chunk should be finalized (buffers returned)
	if atomic.LoadUint32(&ch.freed) != 1 {
		t.Fatalf("chunk buffers were not freed on eviction with refs==0")
	}
	if cs.usedCapB != 0 || cs.usedLenB != 0 {
		t.Fatalf("usage counters should be zero after finalize, got cap=%d len=%d", cs.usedCapB, cs.usedLenB)
	}
}

func TestEvictionDefersUntilRelease(t *testing.T) {
	cs := newTestStore(1<<20, 4)
	stream := uuid.New()
	minSeq := int64(0)
	key0 := ChunkKey{Stream: stream, Index: 0}
	key1 := ChunkKey{Stream: stream, Index: 1}

	ch0 := addChunk(cs, key0, []Frame{makeFrame(cs, 0, 100)})
	ch1 := addChunk(cs, key1, []Frame{makeFrame(cs, 4, 100)})

	// Hold refs by calling GetChunk (LRU hits)
	c0, _ := cs.GetChunk(context.Background(), stream, minSeq, 0)
	c1, _ := cs.GetChunk(context.Background(), stream, minSeq, 4)
	if c0 != ch0 || c1 != ch1 {
		t.Fatal("unexpected chunks returned")
	}

	// Drop limit and evict
	cs.mu.Lock()
	cs.limitB = 0
	cs.evictLocked()
	cs.mu.Unlock()

	// They should be marked evicted but not freed yet
	if atomic.LoadUint32(&ch0.evicted) != 1 || atomic.LoadUint32(&ch1.evicted) != 1 {
		t.Fatalf("expected chunks marked evicted")
	}
	if atomic.LoadUint32(&ch0.freed) == 1 || atomic.LoadUint32(&ch1.freed) == 1 {
		t.Fatalf("chunks should not be freed while refs > 0")
	}

	// After releasing, they should free and counters drop to zero
	cs.ReleaseChunk(c0)
	cs.ReleaseChunk(c1)
	if atomic.LoadUint32(&ch0.freed) != 1 || atomic.LoadUint32(&ch1.freed) != 1 {
		t.Fatalf("expected chunks freed after releases")
	}
	if cs.usedCapB != 0 || cs.usedLenB != 0 {
		t.Fatalf("usage counters should be zero after releasing all, got cap=%d len=%d", cs.usedCapB, cs.usedLenB)
	}
}

func TestPressureGuardBlocksLoads(t *testing.T) {
	cs := newTestStore(1<<20 /*1MiB*/, 4)
	stream := uuid.New()
	minSeq := int64(0)

	// Make cache appear heavily over budget
	cs.mu.Lock()
	cs.usedCapB = cs.limitB*PressureGuardFactor + 1
	cs.mu.Unlock()

	// Miss path (no such key), should error before loadChunk due to guard
	_, err := cs.GetChunk(context.Background(), stream, minSeq, 10)
	if err == nil {
		t.Fatalf("expected cache pressure error, got nil")
	}
}
