# Dependency Injection Examples

## ✅ Correct: Constructor-based DI

```go
// internal/repository/user_repository.go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepository struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepository {
    return &userRepository{db: db}
}

// internal/service/user_service.go
type userService struct {
    repo      repository.UserRepository
    jwtSecret string
}

func NewUserService(repo repository.UserRepository, jwtSecret string) *userService {
    return &userService{
        repo:      repo,
        jwtSecret: jwtSecret,
    }
}

// main.go
repo := repository.NewUserRepository(dbPool)
svc := service.NewUserService(repo, config.JWTSecret)
handler := handler.NewUserHandler(svc)
```


## ❌ Incorrect: Global variables or direct instantiation

```go
// DON'T: Global database connection
var globalDB *pgxpool.Pool

type userService struct{}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) error {
    // DON'T: Using global variable
    _, err := globalDB.Exec(ctx, "INSERT INTO users ...")
    return err
}

// DON'T: Creating dependencies inside
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // DON'T: Direct instantiation
    db, _ := pgxpool.New(context.Background(), "connection-string")
    repo := &userRepository{db: db}
    svc := &userService{repo: repo}
}
```
