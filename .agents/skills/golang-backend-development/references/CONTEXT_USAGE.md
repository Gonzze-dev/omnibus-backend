# Context Usage Examples

## ✅ Correct: Context as first parameter

```go
// internal/service/user_service.go
func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    // Can propagate cancellation
    user, err := s.repo.Create(ctx, user)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// internal/repository/user_repository.go
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    query := `INSERT INTO users (id, email, password) VALUES ($1, $2, $3) RETURNING *`
    
    // Context enables timeout and cancellation
    err := r.db.QueryRow(ctx, query, user.ID, user.Email, user.Password).Scan(
        &user.ID, &user.Email, &user.CreatedAt,
    )
    return user, err
}
```


## ❌ Incorrect: Missing context

```go
// DON'T: No context parameter
func (s *userService) CreateUser(req models.CreateUserRequest) (*models.User, error) {
    // Can't handle cancellation or timeouts
    user, err := s.repo.Create(user)
    return user, err
}

// DON'T: Using background context instead of propagating
func (r *userRepository) Create(user *models.User) (*models.User, error) {
    // DON'T: Creating new context instead of accepting one
    ctx := context.Background()
    err := r.db.QueryRow(ctx, query, user.ID, user.Email).Scan(&user.ID)
    return user, err
}
```

## ❌ Discouraged: Context in Struct

Storing `context.Context` inside a struct is generally an anti-pattern in Go. Contexts are scoped to a request lifecycle, while structs are often long-lived. This pattern creates confusion about the context's lifecycle and makes the code harder to test and maintain.

**Exception:** Some specialized libraries or request-scoped objects (like `http.Request`) might carry a context, but general application structs should not.

```go
// DON'T: Storing context in the struct
type Service struct {
    ctx context.Context // ❌ This ties the service to a single context
    db  *sql.DB
}

func NewService(ctx context.Context, db *sql.DB) *Service {
    return &Service{ctx: ctx, db: db}
}

func (s *Service) DoSomething() error {
    // Uses the stored context, which might be cancelled or expired
    // while the service instance is still alive.
    return s.db.PingContext(s.ctx)
}

// Usage that leads to bugs:
// svc := NewService(ctx, db)
// ... later ...
// svc.DoSomething() // Fails if 'ctx' is already cancelled!
```

