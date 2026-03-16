package main

import "testing"

type Order struct {
	Price float64
	Qty   int64
}

// UpdateInside modifies the object in-place.
func UpdateInside(o *Order, price float64) {
	o.Price = price
}

// UpdateInsideReturn modifies the object in-place and returns it (fluent API style).
func UpdateInsideReturn(o *Order, price float64) *Order {
	return &Order{
		Price: price,
	}
}

// Avoid compiler optimization
var sink *Order

// BenchmarkReturnOnly measures the cost of a simple in-place update.
func BenchmarkReturnOnly(b *testing.B) {
	o := &Order{
		Price: 100,
		Qty:   100,
	}
	b.ResetTimer()
	for i := range b.N {
		UpdateInside(o, float64(i))
	}
}

// BenchmarkReturnModitfy measures the cost of update with a return value (fluent style).
func BenchmarkReturnModitfy(b *testing.B) {
	o := &Order{
		Price: 100,
		Qty:   100,
	}
	b.ResetTimer()
	for i := range b.N {
		sink = UpdateInsideReturn(o, float64(i))
	}
}
