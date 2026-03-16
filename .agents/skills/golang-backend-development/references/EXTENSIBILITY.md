# Extensibility Examples

## ✅ Correct: Event-driven extensibility

```go
// pkg/events/events.go
type Event struct {
    Type      string
    Timestamp time.Time
    Data      interface{}
}

type EventHandler func(Event) error

type EventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        handlers: make(map[string][]EventHandler),
    }
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    handlers := eb.handlers[event.Type]
    eb.mu.RUnlock()
    
    for _, handler := range handlers {
        // Run async to not block main flow
        go func(h EventHandler) {
            if err := h(event); err != nil {
                log.Printf("event handler error: %v", err)
            }
        }(handler)
    }
}

// internal/service/user_service.go
type userService struct {
    repo     repository.UserRepository
    eventBus *events.EventBus
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    user, err := s.repo.Create(ctx, &models.User{
        ID:    uuid.New(),
        Email: req.Email,
    })
    if err != nil {
        return nil, err
    }
    
    // Publish event for side effects
    s.eventBus.Publish(events.Event{
        Type:      "user.created",
        Timestamp: time.Now(),
        Data:      user,
    })
    
    return user, nil
}

// main.go - Wire up side effects without coupling
eventBus := events.NewEventBus()

// Subscribe notification handler
eventBus.Subscribe("user.created", func(e events.Event) error {
    user := e.Data.(*models.User)
    return notificationSvc.SendWelcomeEmail(user.Email)
})

// Subscribe analytics handler
eventBus.Subscribe("user.created", func(e events.Event) error {
    user := e.Data.(*models.User)
    return analyticsSvc.TrackSignup(user.ID)
})

// Subscribe audit log handler
eventBus.Subscribe("user.created", func(e events.Event) error {
    user := e.Data.(*models.User)
    return auditSvc.Log("user_created", user.ID)
})

svc := service.NewUserService(repo, eventBus)
```


## ✅ Correct: Hook pattern for extensibility

```go
// internal/service/hooks.go
type UserHooks struct {
    BeforeCreate []func(context.Context, *models.User) error
    AfterCreate  []func(context.Context, *models.User) error
}

type userService struct {
    repo  repository.UserRepository
    hooks UserHooks
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    user := &models.User{
        ID:    uuid.New(),
        Email: req.Email,
    }
    
    // Run before hooks
    for _, hook := range s.hooks.BeforeCreate {
        if err := hook(ctx, user); err != nil {
            return nil, err
        }
    }
    
    // Core logic
    created, err := s.repo.Create(ctx, user)
    if err != nil {
        return nil, err
    }
    
    // Run after hooks (don't fail the operation if hooks fail)
    for _, hook := range s.hooks.AfterCreate {
        if err := hook(ctx, created); err != nil {
            log.Printf("after create hook error: %v", err)
        }
    }
    
    return created, nil
}

// main.go - Add hooks without modifying service logic
hooks := service.UserHooks{
    BeforeCreate: []func(context.Context, *models.User) error{
        validateUserEmail,
        checkUserQuota,
    },
    AfterCreate: []func(context.Context, *models.User) error{
        sendWelcomeEmail,
        trackAnalytics,
        logAudit,
    },
}

svc := service.NewUserService(repo)
svc.hooks = hooks
```


## ✅ Correct: Proxy Pattern for extensibility

```go
// internal/service/user_service.go
type UserService interface {
    CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
}

// Base implementation
type userService struct {
    repo repository.UserRepository
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    return s.repo.Create(ctx, &models.User{
        ID:    uuid.New(),
        Email: req.Email,
    })
}

// internal/service/proxy/logging_proxy.go
type userLoggingProxy struct {
    next   service.UserService
    logger Logger
}

func NewUserLoggingProxy(next service.UserService, logger Logger) service.UserService {
    return &userLoggingProxy{
        next:   next,
        logger: logger,
    }
}

func (p *userLoggingProxy) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    p.logger.Info("creating user", "email", req.Email)
    
    user, err := p.next.CreateUser(ctx, req)
    
    if err != nil {
        p.logger.Error("failed to create user", "error", err)
    } else {
        p.logger.Info("user created successfully", "id", user.ID)
    }
    
    return user, err
}

// main.go - Chain proxies to add behavior
var svc service.UserService = service.NewUserService(repo)

// Wrap with logging
svc = proxy.NewUserLoggingProxy(svc, logger)

// Wrap with metrics (another proxy)
svc = proxy.NewUserMetricsProxy(svc, metrics)

// The handler receives the fully wrapped service
handler := handler.NewCreateUserHandler(svc)
```


## ❌ Incorrect: Tightly coupled side effects

```go
// DON'T: Hardcoded side effects in business logic
func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    user, err := s.repo.Create(ctx, &models.User{
        ID:    uuid.New(),
        Email: req.Email,
    })
    if err != nil {
        return nil, err
    }
    
    // DON'T: Tightly coupled notification logic
    if err := s.sendWelcomeEmail(user.Email); err != nil {
        log.Printf("failed to send email: %v", err)
    }
    
    // DON'T: Tightly coupled analytics
    if err := s.trackSignup(user.ID); err != nil {
        log.Printf("failed to track: %v", err)
    }
    
    // DON'T: Tightly coupled audit logging
    if err := s.auditLog("user_created", user.ID); err != nil {
        log.Printf("failed to log: %v", err)
    }
    
    // Adding new side effects requires modifying this function
    return user, nil
}

// DON'T: Service depends on too many concerns
type userService struct {
    repo             repository.UserRepository
    emailService     *EmailService
    analyticsService *AnalyticsService
    auditService     *AuditService
    notificationSvc  *NotificationService
    // This list keeps growing...
}
```
