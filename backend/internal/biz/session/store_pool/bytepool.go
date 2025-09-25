package store_pool

import (
	"sort"
	"sync"
)

// ByteBucketPool — пул буферов с фиксированными "вёдрами" по cap.
// JPEG'ы имеют разный размер → чтобы не держать огромный cap у маленьких len,
// используем набор бакетов (32K, 64K, 128K, ...). Это сильно снижает аллокации и давление на GC.
type ByteBucketPool struct {
	sizes []int              // отсортированный список размеров бакетов
	pools map[int]*sync.Pool // key = cap ведра → pool таких буферов
}

func NewByteBucketPool(sizes []int) *ByteBucketPool {
	s := append([]int(nil), sizes...)
	sort.Ints(s)
	pools := make(map[int]*sync.Pool, len(s))
	for _, capSize := range s {
		c := capSize
		pools[c] = &sync.Pool{
			New: func() any { return make([]byte, c) }, // len==cap==c
		}
	}
	return &ByteBucketPool{sizes: s, pools: pools}
}

// BucketSize — подобрать cap-ведро для len=n. 0 — слишком большой буфер (вне пула).
func (bp *ByteBucketPool) BucketSize(n int) int {
	i := sort.SearchInts(bp.sizes, n)
	if i >= len(bp.sizes) {
		return 0
	}
	return bp.sizes[i]
}

// Get — вернуть буфер длиной n (cap — ближайшее ведро >= n). Вне пула — прямая аллокация.
func (bp *ByteBucketPool) Get(n int) []byte {
	if n <= 0 {
		return nil
	}
	if bcap := bp.BucketSize(n); bcap != 0 {
		buf := bp.pools[bcap].Get().([]byte) // len==cap==bcap
		return buf[:n]
	}
	return make([]byte, n) // редкий случай: гигантский кадр
}

// Put — вернуть буфер в пул, только если cap соответствует одному из бакетов.
// Нулить данные не нужно (JPEG не секретный) — экономим CPU.
func (bp *ByteBucketPool) Put(b []byte) {
	if b == nil {
		return
	}
	if p, ok := bp.pools[cap(b)]; ok {
		p.Put(b[:cap(b)]) // возвращаем полный буфер (len==cap)
	}
}
