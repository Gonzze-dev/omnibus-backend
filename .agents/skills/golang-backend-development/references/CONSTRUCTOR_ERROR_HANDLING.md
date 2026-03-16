# Constructor Error Handling

## Overview
Constructors in Go (typically named `NewComponent`) should generally be safe and return errors if initialization fails, rather than panicking. This allows the caller (usually `main.go` or a wireup function) to handle the error gracefully (e.g., logging and exiting, or retrying).

## ✅ Correct: Return Error

If any step in the constructor can fail (e.g., parsing configuration, connecting to a database, compiling a regex), the constructor signature must include `error` as a return value.

```go
package service

import (
    "database/sql"
    "fmt"
)

type DatabaseService struct {
    db *sql.DB
}

// NewDatabaseService returns an error because db.Ping() might fail.
func NewDatabaseService(connStr string) (*DatabaseService, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database connection: %w", err)
    }

    // Explicitly check if the connection is valid
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &DatabaseService{
        db: db,
    }, nil
}
```

## ❌ Incorrect: Panic on Error

Avoid `panic` in constructors. It makes the package hard to use in contexts where the application shouldn't crash (e.g., during tests or if the service is optional).

```go
package service

import (
    "database/sql"
    "log"
)

type DatabaseService struct {
    db *sql.DB
}

// DON'T: Panicking hides control flow and crashes the app immediately.
func NewDatabaseService(connStr string) *DatabaseService {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        // ❌ Avoid panic
        panic(err)
    }

    if err := db.Ping(); err != nil {
        // ❌ Avoid log.Fatal (which calls exit) inside libraries
        log.Fatalf("failed to ping db: %v", err)
    }

    return &DatabaseService{
        db: db,
    }
}
```
