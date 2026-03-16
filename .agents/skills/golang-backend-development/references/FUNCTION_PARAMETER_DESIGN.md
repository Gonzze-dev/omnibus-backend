# Function Parameter Design Examples

## ✅ Correct: Functional Options Pattern

```go
// internal/service/user_service.go
type CreateUserOptions struct {
    SendWelcomeEmail bool
    AssignRole       string
    Metadata         map[string]string
}

type CreateUserOption func(*CreateUserOptions)

func WithWelcomeEmail() CreateUserOption {
    return func(o *CreateUserOptions) {
        o.SendWelcomeEmail = true
    }
}

func WithRole(role string) CreateUserOption {
    return func(o *CreateUserOptions) {
        o.AssignRole = role
    }
}

func WithMetadata(metadata map[string]string) CreateUserOption {
    return func(o *CreateUserOptions) {
        o.Metadata = metadata
    }
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest, opts ...CreateUserOption) (*models.User, error) {
    options := &CreateUserOptions{
        SendWelcomeEmail: false,
        AssignRole:       "user",
    }
    
    for _, opt := range opts {
        opt(options)
    }
    
    // Use options...
    user, err := s.repo.Create(ctx, &models.User{
        Email: req.Email,
        Role:  options.AssignRole,
    })
    
    if err != nil {
        return nil, err
    }
    
    if options.SendWelcomeEmail {
        s.sendWelcomeEmail(ctx, user)
    }
    
    return user, nil
}

// Usage - backward compatible as new options are added
user, err := svc.CreateUser(ctx, req)
user, err := svc.CreateUser(ctx, req, WithWelcomeEmail())
user, err := svc.CreateUser(ctx, req, WithWelcomeEmail(), WithRole("admin"))
```


## ✅ Correct: Options Struct Pattern

```go
type ListUsersRequest struct {
    Limit      int
    Offset     int
    SortBy     string
    SortOrder  string
    FilterRole string
}

func (s *userService) ListUsers(ctx context.Context, req ListUsersRequest) ([]*models.User, error) {
    // Set defaults
    if req.Limit == 0 {
        req.Limit = 20
    }
    if req.SortBy == "" {
        req.SortBy = "created_at"
    }
    if req.SortOrder == "" {
        req.SortOrder = "DESC"
    }
    
    return s.repo.List(ctx, req)
}

// Easy to extend without breaking existing calls
// Before: ListUsers(ctx, 20, 0)
// After: ListUsers(ctx, ListUsersRequest{Limit: 20, Offset: 0})
// Later: ListUsers(ctx, ListUsersRequest{Limit: 20, Offset: 0, FilterRole: "admin"})
```


## ❌ Incorrect: Multiple positional parameters

```go
// DON'T: Too many positional parameters - not extensible
func (s *userService) CreateUser(
    ctx context.Context,
    email string,
    password string,
    sendWelcomeEmail bool,
    role string,
    metadata map[string]string,
) (*models.User, error) {
    // Adding new parameters breaks all existing calls
}

// DON'T: Changing function signature breaks compatibility
// Version 1
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)

// Version 2 - BREAKS existing code
func (s *userService) ListUsers(ctx context.Context, limit, offset int, sortBy string) ([]*models.User, error)
```
