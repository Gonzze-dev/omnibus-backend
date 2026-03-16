# Design Principles Examples

## ✅ Correct: Simple and focused

```go
// Single Responsibility - one clear purpose
type EmailSender struct {
    smtpHost string
    smtpPort int
}

func (e *EmailSender) Send(to, subject, body string) error {
    // Simple, focused implementation
    return smtp.SendMail(
        fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort),
        nil,
        "noreply@example.com",
        []string{to},
        []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body)),
    )
}

// Interface Segregation - small, focused interface
type UserCreator interface {
    CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
}

// Dependency Inversion - depend on abstraction
type NotificationService struct {
    emailSender EmailSender // Could be interface if needed
}
```


## ❌ Incorrect: Over-engineered

```go
// DON'T: Over-abstraction without clear benefit
type AbstractFactoryProvider interface {
    GetFactory() AbstractFactory
}

type AbstractFactory interface {
    CreateUserFactory() UserFactory
}

type UserFactory interface {
    CreateUser() User
}

// DON'T: God object with too many responsibilities
type UserManager struct {
    db            *pgxpool.Pool
    cache         *redis.Client
    emailSender   *EmailSender
    logger        *Logger
    metrics       *MetricsCollector
    eventBus      *EventBus
}

func (m *UserManager) CreateUser() error { /* ... */ }
func (m *UserManager) UpdateUser() error { /* ... */ }
func (m *UserManager) DeleteUser() error { /* ... */ }
func (m *UserManager) SendEmail() error { /* ... */ }
func (m *UserManager) LogActivity() error { /* ... */ }
func (m *UserManager) PublishEvent() error { /* ... */ }
```

