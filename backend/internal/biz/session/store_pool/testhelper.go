package store_pool

// InjectChunkForTest Test-only helper to let other packages (httpapi tests) prefill the LRU.
// This is compiled only during `go test` due to the build tag above.
func InjectChunkForTest(key ChunkKey, ch *Chunk, cs *ChunkStore) {
	cs.mu.Lock()
	el := &lruEntry{key: key, chunk: ch}
	le := cs.lru.PushFront(el)
	cs.items[key] = le
	cs.usedLenB += ch.BytesLen
	cs.usedCapB += ch.BytesCap
	cs.mu.Unlock()
}
