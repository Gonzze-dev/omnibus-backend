# Project Organization Examples

## ✅ Correct: Slim main.go

```go
// backend/cmd/api/main.go
package main

import (
    "context"
    "log"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    "yourproject/internal/handler"
    "yourproject/internal/repository"
    "yourproject/internal/service"
)

func main() {
    dbPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer dbPool.Close()

    // Dependency injection
    repo := repository.NewRepository(dbPool)
    svc := service.NewService(repo, os.Getenv("JWT_SECRET"))
    h := handler.NewHandler(svc)

    // Start server
    if err := h.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}
```


## ❌ Incorrect: Business logic in main.go

```go
// backend/cmd/api/main.go
package main

import (
    "encoding/json"
    "net/http"
)

func main() {
    // DON'T: Business logic in main.go
    http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
        
        var user struct {
            Email string `json:"email"`
        }
        json.NewDecoder(r.Body).Decode(&user)
        
        // Business logic should be in service layer
        if user.Email == "" {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        // ... more logic
    })
    
    http.ListenAndServe(":8080", nil)
}
```
