# Goroutine Pools & Avoiding Memory Leaks

## The Problem: Unbounded Goroutines

Go's lightweight goroutines are powerful, but creating them without bound can lead to serious issues in high-throughput applications:
- **Memory Leaks/OOM**: Each goroutine consumes stack memory (initially ~2KB, but can grow). Spawning millions of goroutines (e.g., one per request in a massive traffic spike) can exhaust system RAM.
- **Scheduler Thrashing**: Excessive goroutines increase GC pressure and scheduling overhead, degrading CPU performance.
- **File Descriptor Exhaustion**: If goroutines hold resources (like sockets or DB connections), you may hit OS limits.

See [Go Issue #9869](https://github.com/golang/go/issues/9869) for technical context on goroutine memory issues.

## The Solution: Worker Pools

When you have a task that generates a high volume of concurrent work (e.g., a high-traffic HTTP handler that spawns a background task for every request), you **must** use a **Goroutine Pool** to limit concurrency and reuse goroutines.

Recommended Libraries:
- [ants](https://github.com/panjf2000/ants) (High performance, low memory footprint, auto-scaling)
- [tunny](https://github.com/Jeffail/tunny) (Fixed-size worker pool, good for CPU-bound tasks)

## Example: Using `ants`

`ants` is an efficient goroutine pool that automatically manages and recycles goroutines.

### ❌ Bad Practice (Unbounded)

```go
func (s *Service) ProcessBatch(items []Item) {
    for _, item := range items {
        // DANGEROUS: If items has 100k elements, this spawns 100k goroutines immediately.
        // This can crash the runtime or OOM the container.
        go s.process(item)
    }
}
```

### ✅ Good Practice (With `ants` Pool)

```go
import (
    "log"
    "sync"
    "github.com/panjf2000/ants/v2"
)

func (s *Service) ProcessBatch(items []Item) {
    var wg sync.WaitGroup

    // Option 1: Use the default common pool (simple, shared)
    // defer ants.Release() // usually called at app shutdown

    // Option 2: Instantiate a specific pool for this task (recommended for isolation)
    // Limits concurrency to 1000 active goroutines
    pool, _ := ants.NewPool(1000)
    defer pool.Release()

    for _, item := range items {
        wg.Add(1)
        // Submit the task to the pool
        err := pool.Submit(func() {
            defer wg.Done()
            s.process(item)
        })
        
        if err != nil {
            // Handle pool overload (e.g., queue full) or other errors
            log.Printf("failed to submit task: %v", err)
            wg.Done() 
        }
    }

    wg.Wait()
}
```

## Example: Using `tunny`

`tunny` is great when you need a fixed number of workers constantly processing a stream of jobs.

```go
import (
    "github.com/Jeffail/tunny"
    "runtime"
)

func main() {
    numCPUs := runtime.NumCPU()
    
    // Create a pool with a worker count equal to CPU cores
    pool := tunny.NewFunc(numCPUs, func(payload interface{}) interface{} {
        // Assert payload type
        val := payload.(int)
        return val * 2
    })
    defer pool.Close()

    // Submit work
    result := pool.Process(10)
    fmt.Println(result)
}
```

## Handling Infinite Loops in Goroutines

When a goroutine is created to run a background task involving an infinite loop (e.g., polling, processing a stream, or a ticker), you **must** ensure it can be terminated gracefully. Failing to do so results in "goroutine leaks," where the goroutine continues running even after the parent process or request has finished.

### ❌ Bad Practice (Uncancellable Loop)

```go
func (s *Service) StartBackgroundWorker() {
    go func() {
        for {
            // This loop will run forever until the application crashes or exits.
            // There is no way to stop it programmatically.
            s.doWork()
            time.Sleep(1 * time.Second)
        }
    }()
}
```

### ✅ Good Practice (Context-Aware Loop)

Always pass a `context.Context` to control the lifecycle of the goroutine.

```go
func (s *Service) StartBackgroundWorker(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                // Context cancelled (timeout, request finished, or shutdown signal)
                // Clean up and exit the goroutine
                log.Println("Stopping background worker:", ctx.Err())
                return
            case <-ticker.C:
                // Perform the work
                s.doWork(ctx)
            }
        }
    }()
}

// Usage Example
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    svc := &Service{}
    
    svc.StartBackgroundWorker(ctx)
    
    // ... run application ...
    
    // When shutting down
    cancel() // This signals the goroutine to stop
    // Wait for cleanup if necessary
}
```

## Channel Ownership and Closure

To prevent panics (writing to a closed channel) and deadlocks, follow the strict rule: **The goroutine that writes to the channel is responsible for closing it.**

### ✅ Good Practice (Writer Closes)

```go
func (s *Service) StreamData(ctx context.Context) <-chan Data {
    ch := make(chan Data)

    go func() {
        defer close(ch) // Writer closes the channel when done

        for {
            data, err := s.fetchNext(ctx)
            if err != nil {
                return // Exit, defer will close channel
            }
            
            select {
            case ch <- data:
                // Data sent
            case <-ctx.Done():
                return // Exit, defer will close channel
            }
        }
    }()

    return ch
}

func Consumer(ch <-chan Data) {
    // Consumer just ranges over the channel until it's closed by the writer
    for data := range ch {
        process(data)
    }
}
```

### ❌ Bad Practice (Receiver Closes or Ambiguous Ownership)

```go
func (s *Service) BadStream() chan Data {
    ch := make(chan Data)
    go func() {
        // Writer does NOT close channel
        ch <- Data{}
    }()
    return ch
}

func Consumer(ch chan Data) {
    data := <-ch
    process(data)
    close(ch) // DANGEROUS: If the writer tries to send more, it will panic!
}
```
