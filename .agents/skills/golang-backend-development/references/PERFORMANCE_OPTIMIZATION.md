# Performance Optimization Examples

In scenarios with extreme performance requirements, use memory management and allocation strategies to reduce Garbage Collection (GC) overhead.

## 1. Using `sync.Pool` for Frequent Allocations

When objects are created and destroyed at a high frequency (e.g., processing thousands of requests per second), reusing objects can significantly reduce GC pressure.

```go
// internal/pkg/buffer/pool.go
var bufferPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) {
    // Get a buffer from the pool
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.Write(data)
    // Perform operations with buf...
}
```

> [!NOTE]
> Reference: [pool_test.go](file:///Users/septemlee/.gemini/skills/golang-backend/benchmarks/pool_test.go)

## 2. Return-by-Value to Reduce GC Pressure

Returning a concrete object rather than a pointer often allows the Go compiler to perform **escape analysis** and allocate the object on the **stack** instead of the **heap**. Stack allocation is much cheaper and doesn't require GC.

### ❌ Less Optimal: Always returning pointers

```go
func NewUser(name string) *User {
    return &User{Name: name} // Likely escapes to heap
}
```

### ✅ More Optimal: Returning by value (for small structs)

```go
func NewUser(name string) User {
    return User{Name: name} // Often stays on stack
}
```

> [!NOTE]
> Reference: [allocation_test.go](file:///Users/septemlee/.gemini/skills/golang-backend/benchmarks/allocation_test.go)

## 3. In-place Mutation vs. Copying (Hot Paths Only)

While `SKILL.md` generally forbids mutating input parameters for predictability, in **extreme hot paths**, in-place mutation can be used to avoid the overhead of copying or allocating new structs.

### ❌ Standard: Copy and Return

```go
func UpdateOrder(o Order, price float64) Order {
    // Operates on a copy
    return Order {
        Price: price,
        ...
    }
}
```

### ✅ Performance Optimized: In-place Mutation

```go
func UpdateOrderInside(o *Order, price float64) {
    o.Price = price // Direct modification, no allocation
}
```

> [!NOTE]
> Reference: [mutation_test.go](file:///Users/septemlee/.gemini/skills/golang-backend/benchmarks/mutation_test.go)

### Rationale: Stack vs Heap
- **Stack**: Automatically cleaned up when the function returns. Extremely fast.
- **Heap**: Requires the GC to track and clean up. Adds latency and CPU overhead.

## When to Optimize
> [!WARNING]
> Only apply these patterns when there is a documented performance bottleneck. Premature optimization can lead to more complex code with subtle bugs (e.g., using an object after returning it to a `sync.Pool`).

1. **High-Frequency Hot Paths**: Code triggered thousands of times per second.
2. **Memory Pressured Environments**: Where reducing object churn helps maintain low latency.
3. **Profiling Evidence**: Use `go test -bench` and `pprof` to justify the optimization.

## Comparative Benchmarks

Detailed performance comparisons can be found in the `benchmarks/` directory:

- **Object Pooling**: [pool_test.go](../benchmarks/pool_test.go) - Compares `sync.Pool` vs. frequent heap allocations.
- **Allocation Strategies**: [allocation_test.go](../benchmarks/allocation_test.go) - Compares Return-by-Pointer (Heap) vs. Return-by-Value (Stack).
- **Mutation Patterns**: [mutation_test.go](../benchmarks/mutation_test.go) - Compares In-place Mutation vs. Copy-and-Return.
