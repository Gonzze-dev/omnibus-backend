# Testing & Coverage Examples

## ✅ Correct: Testable code with mocks

```go
// internal/service/user_service_test.go
//go:generate mockery --name=UserRepository --output=./mocks --outpkg=mocks

func TestUserService_CreateUser(t *testing.T) {
    mockRepo := new(mocks.UserRepository)
    svc := NewUserService(mockRepo, "test-secret")
    
    req := models.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    expectedUser := &models.User{
        ID:    uuid.New(),
        Email: req.Email,
    }
    
    // Setup mock expectation
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
        Return(expectedUser, nil)
    
    // Execute
    user, err := svc.CreateUser(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Email, user.Email)
    mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_ValidationError(t *testing.T) {
    mockRepo := new(mocks.UserRepository)
    svc := NewUserService(mockRepo, "test-secret")
    
    req := models.CreateUserRequest{
        Email:    "invalid-email",
        Password: "123",
    }
    
    // Execute
    user, err := svc.CreateUser(context.Background(), req)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, ErrInvalidEmail, err)
}

// internal/repository/user_repository_test.go
func TestUserRepository_Create(t *testing.T) {
    // Use testcontainers or pgx test pool
    ctx := context.Background()
    pool := setupTestDB(t)
    defer pool.Close()
    
    repo := NewUserRepository(pool)
    
    user := &models.User{
        ID:       uuid.New(),
        Email:    "test@example.com",
        Password: "hashed-password",
    }
    
    created, err := repo.Create(ctx, user)
    
    assert.NoError(t, err)
    assert.NotNil(t, created)
    assert.Equal(t, user.Email, created.Email)
}
```


## ❌ Incorrect: Untestable code

```go
// DON'T: Direct database access without interface
type userService struct {
    db *pgxpool.Pool // Can't mock this easily
}

func (s *userService) CreateUser(req models.CreateUserRequest) error {
    // DON'T: Hard to test - direct DB access
    _, err := s.db.Exec(context.Background(), "INSERT INTO users ...")
    return err
}

// DON'T: Global state makes testing difficult
var globalConfig = loadConfig()

func CreateUser(email string) error {
    // Uses global state - hard to test
    db := connectDB(globalConfig.DBUrl)
    // ...
}
```
