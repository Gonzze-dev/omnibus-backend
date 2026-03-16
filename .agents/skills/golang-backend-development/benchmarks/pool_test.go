package main

import (
	"sync"
	"testing"
)

// BigObject simulates a significant memory footprint (approx 8KB).
type BigObject struct {
	ID       int64
	Name     string
	Data     [1024]int64 // 1024 * 8 bytes = 8KB
	IsActive bool
}

var globalSinkPool *BigObject

var objPool = sync.Pool{
	New: func() any {
		return new(BigObject)
	},
}

// BenchmarkStructWithoutPool measures the performance of frequent heap allocations.
func BenchmarkStructWithoutPool(b *testing.B) {
	for i := range b.N {
		// New allocation every iteration, likely escaping to heap
		obj := &BigObject{
			ID: int64(i),
		}
		globalSinkPool = obj
	}
}

// BenchmarkStructWithPool measures the performance of reusing objects via sync.Pool.
func BenchmarkStructWithPool(b *testing.B) {
	for i := range b.N {
		// Acquire from pool
		obj := objPool.Get().(*BigObject)

		// Reset state (Critical step when using sync.Pool)
		obj.ID = int64(i)
		obj.IsActive = true

		globalSinkPool = obj

		// Release back to pool
		objPool.Put(obj)
	}
}
