# Avoid Mutating Parameters Examples

Mutating input parameters (especially pointers) inside a function can lead to unexpected side effects at the call-site, making the code harder to reason about. This should be avoided unless there is a critical performance justification.

## ❌ Incorrect: Mutating a Pointer Parameter

The caller might not expect the original object to be modified, leading to bugs when the object is reused.

```go
// internal/service/user_service.go
func (s *userService) UpdateStatus(user *models.User, status string) {
    // DON'T: Mutating the input pointer directly
    user.Status = status
    user.UpdatedAt = time.Now()
    
    s.repo.Update(user)
}

// Call-site
user := &models.User{ID: 1, Name: "Alice", Status: "active"}
svc.UpdateStatus(user, "inactive")
// user.Status is now "inactive" here, which might be unexpected if the caller
// planned to use the original "active" state later.
```

## ✅ Correct: Returning a New Value or Using Local Copy

Return the modified object or a new state to make the change explicit to the caller.

```go
// internal/service/user_service.go
func (s *userService) UpdateStatus(ctx context.Context, user *models.User, status string) (*models.User, error) {
    // DO: Create a copy or return the result
    updatedUser := *user // Shallow copy
    updatedUser.Status = status
    updatedUser.UpdatedAt = time.Now()
    
    if err := s.repo.Update(ctx, &updatedUser); err != nil {
        return nil, err
    }
    
    return &updatedUser, nil
}

// Call-site
user := &models.User{ID: 1, Name: "Alice", Status: "active"}
updatedUser, err := svc.UpdateStatus(ctx, user, "inactive")
// The original 'user' remains "active", while 'updatedUser' is "inactive".
```

## When to Consider Mutation (Performance Exception)

In most business logic, the overhead of copying a struct is negligible. However, you may consider mutation in the following cases:

1. **High-Frequency Hot Paths**: In performance-critical loops where allocations or large copies significantly impact latency or CPU usage.
2. **Extremely Large Structs**: If copying the struct would consume excessive memory or time.
3. **Low-Level Libraries**: Where performance is the primary goal and the API clearly documents the mutation behavior.

> [!IMPORTANT]
> Always prioritize readability and predictability. Only opt for mutation if profiling data proves a significant performance benefit.

## Rationale
1. **Predictability**: The caller's data remains unchanged unless they explicitly use the return value.
2. **Context Understanding**: Reduces the cognitive load of tracking how a variable changes across multiple function calls.
3. **Concurrency Safety**: Reduces the risk of race conditions when the same object is passed to multiple goroutines.
