# Layered Responsibility Examples

## ✅ Correct: Proper Layer Separation

```go
// internal/handler/user_handler.go
type UserHandler struct {
    userService service.UserService // Accept interface
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req models.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    // Defer to service - no business logic here
    user, err := h.userService.CreateUser(r.Context(), req)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusCreated, user)
}

// internal/service/user_service.go
type UserService interface {
    CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
}

type userService struct {
    repo repository.UserRepository
}

// Return concrete type, accept interface
func NewUserService(repo repository.UserRepository) *userService {
    return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    // Business logic here
    if err := validateEmail(req.Email); err != nil {
        return nil, ErrInvalidEmail
    }

    hashedPassword, err := hashPassword(req.Password)
    if err != nil {
        return nil, err
    }

    user := &models.User{
        ID:       uuid.New(),
        Email:    req.Email,
        Password: hashedPassword,
    }

    return s.repo.Create(ctx, user)
}
```


## ❌ Incorrect: Business logic in handler

```go
// internal/handler/user_handler.go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req models.CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)

    // DON'T: Validation and business logic in handler
    if !strings.Contains(req.Email, "@") {
        respondError(w, http.StatusBadRequest, errors.New("invalid email"))
        return
    }

    if len(req.Password) < 8 {
        respondError(w, http.StatusBadRequest, errors.New("password too short"))
        return
    }

    // DON'T: Direct database access from handler
    hashedPassword := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
    _, err := h.db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", 
        req.Email, hashedPassword)
}
```


## ✅ Correct: Using Handler Wrapper

```go
// internal/handler/base.go
type Handler[REQ, RESP any] struct {
    HandleFunc    func(ctx context.Context, req REQ) (RESP, error)
    DecodeFunc    func(r *http.Request) (REQ, error)
    EncodeFunc    func(w http.ResponseWriter, resp RESP) error
    ErrorHandler  func(w http.ResponseWriter, err error)
}

func (h Handler[REQ, RESP]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    req, err := h.DecodeFunc(r)
    if err != nil {
        h.ErrorHandler(w, err)
        return
    }

    resp, err := h.HandleFunc(r.Context(), req)
    if err != nil {
        h.ErrorHandler(w, err)
        return
    }

    if err := h.EncodeFunc(w, resp); err != nil {
        h.ErrorHandler(w, err)
    }
}

// internal/handler/user_handler.go
func NewCreateUserHandler(svc service.UserService) http.Handler {
    return Handler[models.CreateUserRequest, models.User]{
        HandleFunc: func(ctx context.Context, req models.CreateUserRequest) (models.User, error) {
            return svc.CreateUser(ctx, req)
        },
        DecodeFunc:   DefaultDecoder[models.CreateUserRequest],
        EncodeFunc:   DefaultEncoder[models.User],
        ErrorHandler: DefaultErrorHandler,
    }
}
```


## ❌ Incorrect: Manual request/response handling

```go
// DON'T: Reimplementing the same logic in every handler
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req models.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    user, err := h.svc.CreateUser(r.Context(), req)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```
