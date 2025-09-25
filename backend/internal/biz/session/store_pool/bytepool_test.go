package store_pool

import "testing"

func TestBucketSizeIncreasing(t *testing.T) {
	sizes := []int{32 << 10, 64 << 10, 128 << 10}
	bp := NewByteBucketPool(sizes)
	tests := []struct{ n, want int }{
		{1, 32 << 10},
		{32<<10 - 1, 32 << 10},
		{32<<10 + 1, 64 << 10},
		{64<<10 + 1, 128 << 10},
		{129 << 10, 0}, // вне пула
	}
	for _, tc := range tests {
		got := bp.BucketSize(tc.n)
		if got != tc.want {
			t.Fatalf("BucketSize(%d)=%d want=%d", tc.n, got, tc.want)
		}
	}
}

func TestGetPutRoundTrip(t *testing.T) {
	bp := NewByteBucketPool([]int{32 << 10, 64 << 10})
	b := bp.Get(100) // должен прийти буфер из ведра 32К
	if len(b) != 100 || cap(b) != 32<<10 {
		t.Fatalf("unexpected len/cap: len=%d cap=%d", len(b), cap(b))
	}
	for i := 0; i < len(b); i++ {
		b[i] = byte(i)
	}
	bp.Put(b)
	// второй Get того же размера снова должен дать ведро 32К
	b2 := bp.Get(200)
	if len(b2) != 200 || cap(b2) != 32<<10 {
		t.Fatalf("unexpected len/cap on second get: len=%d cap=%d", len(b2), cap(b2))
	}
}

func TestHugeOutsidePool(t *testing.T) {
	bp := NewByteBucketPool([]int{32 << 10})
	b := bp.Get(1<<20 + 123)
	if cap(b) < len(b) {
		t.Fatal("cap should be >= len for huge alloc")
	}
	bp.Put(b) // не должен паниковать, но и в пул не кладётся
}
