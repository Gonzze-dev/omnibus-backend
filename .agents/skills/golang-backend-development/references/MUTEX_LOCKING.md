# Mutex Locking

When dealing with concurrent access to shared resources, minimizing the duration that a mutex lock is held is critical for performance. 

## The Problem: Coarse-Grained Locking

Using a single mutex for multiple independent resources (coarse-grained locking) can lead to unnecessary contention. If one goroutine holds the lock to access resource A, another goroutine trying to access resource B must wait, even though the resources are unrelated.

```go
type Service struct {
    mu             sync.Mutex
    failCounter    int
    successCounter int
}

func (s *Service) IncrementFail() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.failCounter++
}

func (s *Service) IncrementSuccess() {
    s.mu.Lock()
    defer s.mu.Unlock() // Blocks IncrementFail unnecessarily
    s.successCounter++
}
```

## The Solution: Fine-Grained Locking with `Mutex[T]`

To avoid this, each independent resource should have its own mutex. We can use a generic `Mutex[T]` wrapper to enforce this pattern and ensure thread safety for individual fields.

### `pkg/syncutil/mutex.go`

```go
package syncutil

import "sync"

// Mutex wraps a value and protects it with a sync.Mutex.
// This ensures that access to the value is always synchronized.
type Mutex[T any] struct {
    mu sync.Mutex
    v  T
}

// NewMutex creates a new Mutex[T] with the initial value.
func NewMutex[T any](initial T) *Mutex[T] {
    return &Mutex[T]{
        v: initial,
    }
}

// Do executes the given function while holding the lock.
// The function receives the current value and returns the new value.
func (m *Mutex[T]) Do(f func(v T) T) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.v = f(m.v)
}
```

### Usage Example: `internal/service/service.go`

```go
package service

import (
    "context"
    
    "your-project/pkg/syncutil"
)

type Service struct {
    // Each counter has its own independent lock.
    failCounter    *syncutil.Mutex[int]
    successCounter *syncutil.Mutex[int]
}

func NewService() *Service {
    return &Service{
        failCounter:    syncutil.NewMutex(0),
        successCounter: syncutil.NewMutex(0),
    }
}

func (s *Service) RecordFailure(ctx context.Context) {
    s.failCounter.Do(func(v int) int {
        return v + 1
    })
}

func (s *Service) RecordSuccess(ctx context.Context) {
    // This will not block or be blocked by RecordFailure
    s.successCounter.Do(func(v int) int {
        return v + 1
    })
}
```

## Benefits

1.  **Reduced Contention**: Operations on `failCounter` do not block operations on `successCounter`.
2.  **Encapsulation**: The mutex and the data it protects are bundled together. You cannot accidentally access the data without the lock (unless you bypass the wrapper, which is harder to do accidentally).
3.  **Clarity**: It is immediately obvious which lock protects which piece of data.
