package main

import "testing"

// SmallObject is a compact struct that should stay on the stack when returned by value.
type SmallObject struct {
	ID   int64
	Type int
}

var (
	sinkPointer  *SmallObject
	sinkConcrete SmallObject
)

// ReturnPointer allocates on the heap (due to returning address of local).
//
//go:noinline
func ReturnPointer(id int64) *SmallObject {
	return &SmallObject{ID: id, Type: 1}
}

// ReturnConcrete returns by value, allowing the compiler to keep it on the stack.
//
//go:noinline
func ReturnConcrete(id int64) SmallObject {
	return SmallObject{ID: id, Type: 1}
}

// BenchmarkGetPointer measures performance when objects escape to the heap.
func BenchmarkGetPointer(b *testing.B) {
	for i := range b.N {
		sinkPointer = ReturnPointer(int64(i))
	}
}

// BenchmarkGetConcrete measures performance when objects stay on the stack.
func BenchmarkGetConcrete(b *testing.B) {
	for i := range b.N {
		sinkConcrete = ReturnConcrete(int64(i))
	}
}
