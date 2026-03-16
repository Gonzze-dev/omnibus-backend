# Must-Prefix Functions & Panic Safety

## Overview
Must-prefix functions and panic helpers are utilities that panic on error instead of returning errors. While convenient for initialization, they must be used carefully to prevent runtime crashes in production services.

## Core Principle
**Must-functions should only be used where immediate failure is acceptable and expected, or where inputs are guaranteed to be valid.**

---

## Safe Usage Patterns

### 1. Initialization in main.go

Must-functions are acceptable during application startup where failure should terminate the process.

```go
// main.go
package main

import (
    "database/sql"
    "log"
    "regexp"
    "text/template"
)

// Helper function for initialization
func Must[T any](t T, err error) T {
    if err != nil {
        panic(err)
    }
    return t
}

var (
    // ✅ Safe: Compile-time constant, app cannot function without valid regex
    emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
    
    // ✅ Safe: Static template loaded at startup
    welcomeTemplate = template.Must(template.ParseFiles("templates/welcome.html"))
)

func main() {
    // ✅ Safe: Critical dependency, app should fail fast if DB unavailable
    db := Must(sql.Open("postgres", "postgres://localhost/mydb"))
    defer db.Close()
    
    // ✅ Safe: Required config, no point running without it
    cfg := Must(loadConfig("config.yaml"))
    
    log.Println("Server starting...")
    // ... rest of initialization
}
```

### 2. Guaranteed Safe Inputs

Use Must-functions only when you can manually verify inputs will never fail.

```go
package validator

import "regexp"

// ✅ Safe: Hardcoded pattern is verified to be valid
var (
    uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
    phonePattern = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

func IsValidUUID(s string) bool {
    return uuidPattern.MatchString(s)
}
```

---

## Unsafe Usage Patterns - Anti-Patterns

### ❌ Anti-Pattern 1: User Input Validation

**WRONG - Handler with Must-function:**
```go
// ❌ DANGEROUS: User can send invalid regex and crash the service
func (h *UserHandler) SearchByPattern(w http.ResponseWriter, r *http.Request) {
    pattern := r.URL.Query().Get("pattern")
    
    // This will PANIC if pattern is invalid regex!
    regex := regexp.MustCompile(pattern)
    
    results := h.service.SearchUsers(r.Context(), regex)
    json.NewEncoder(w).Encode(results)
}
```

**CORRECT - Proper Error Handling:**
```go
// ✅ CORRECT: Returns error for invalid input
func (h *UserHandler) SearchByPattern(w http.ResponseWriter, r *http.Request) {
    pattern := r.URL.Query().Get("pattern")
    
    regex, err := regexp.Compile(pattern)
    if err != nil {
        http.Error(w, "invalid regex pattern", http.StatusBadRequest)
        return
    }
    
    results := h.service.SearchUsers(r.Context(), regex)
    json.NewEncoder(w).Encode(results)
}
```

### ❌ Anti-Pattern 2: Dynamic Template Loading

**WRONG - Runtime Template Compilation:**
```go
// ❌ DANGEROUS: External template could be malformed
func (s *EmailService) SendCustomEmail(ctx context.Context, userID uuid.UUID, templatePath string) error {
    // This will PANIC if template file is invalid!
    tmpl := template.Must(template.ParseFiles(templatePath))
    
    return s.sendEmail(ctx, userID, tmpl)
}
```

**CORRECT - Handle Template Errors:**
```go
// ✅ CORRECT: Returns error for invalid template
func (s *EmailService) SendCustomEmail(ctx context.Context, userID uuid.UUID, templatePath string) error {
    tmpl, err := template.ParseFiles(templatePath)
    if err != nil {
        return fmt.Errorf("failed to parse template: %w", err)
    }
    
    return s.sendEmail(ctx, userID, tmpl)
}
```

### ❌ Anti-Pattern 3: Database Operations

**WRONG - Must in Service Layer:**
```go
// ❌ DANGEROUS: DB errors should be handled gracefully
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // This will PANIC on DB connection issues!
    result := Must(s.repo.Insert(ctx, req))
    
    return result, nil
}
```

**CORRECT - Proper Error Propagation:**
```go
// ✅ CORRECT: Propagates errors to caller
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    user, err := s.repo.Insert(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

### ❌ Anti-Pattern 4: Configuration Loading in Services

**WRONG - Runtime Config Parsing:**
```go
// ❌ DANGEROUS: Config changes could crash running service
func (s *FeatureService) IsEnabled(feature string) bool {
    cfg := Must(s.loadFeatureConfig()) // Panics if config file is corrupted!
    return cfg.Features[feature]
}
```

**CORRECT - Handle Config Errors:**
```go
// ✅ CORRECT: Returns error, allows service to continue
func (s *FeatureService) IsEnabled(feature string) (bool, error) {
    cfg, err := s.loadFeatureConfig()
    if err != nil {
        return false, fmt.Errorf("failed to load config: %w", err)
    }
    return cfg.Features[feature], nil
}
```

---

## Generic Must Helper Implementation

If you implement a generic `Must` helper, place it in a utility package with clear documentation:

```go
// pkg/must/must.go
package must

// Must is a helper that panics if err is not nil, otherwise returns t.
// WARNING: Only use during application initialization (e.g., in main.go).
// NEVER use in request handlers, services, or repositories.
//
// Example safe usage:
//   db := must.Must(sql.Open("postgres", dsn))  // in main.go
//
// Example unsafe usage:
//   user := must.Must(repo.GetUser(ctx, id))    // in handler - DO NOT DO THIS
func Must[T any](t T, err error) T {
    if err != nil {
        panic(err)
    }
    return t
}

// MustValue is like Must but for functions that only return a value.
func MustValue[T any](fn func() (T, error)) T {
    t, err := fn()
    if err != nil {
        panic(err)
    }
    return t
}
```

---

## Decision Tree

Use this decision tree to determine if Must-functions are appropriate:

```
Is this code in main.go or init()?
├─ YES: Is the failure acceptable/expected at startup?
│  ├─ YES: ✅ Must-function is acceptable
│  └─ NO: ❌ Use proper error handling
└─ NO: Is this a compile-time constant or hardcoded value?
   ├─ YES: Have you manually verified it cannot fail?
   │  ├─ YES: ✅ Must-function might be acceptable
   │  └─ NO: ❌ Use proper error handling
   └─ NO: Does this involve user input, external data, or runtime values?
      └─ YES: ❌ NEVER use Must-function
```

---

## Summary

### ✅ Safe Contexts for Must-Functions:
1. Application initialization in `main.go`
2. Package-level variable initialization with constants
3. Hardcoded values that are manually verified
4. Critical dependencies where app cannot function without them

### ❌ Unsafe Contexts - NEVER Use Must-Functions:
1. HTTP handlers or middleware
2. Service layer methods
3. Repository operations
4. Any code processing user input
5. Dynamic file/template loading
6. Runtime configuration changes
7. Database operations
8. External API calls

### Why This Matters:
- **Production Stability**: Panics crash the entire service, affecting all users
- **Observability**: Errors can be logged, metrics collected, and alerts triggered
- **Graceful Degradation**: Errors allow fallback behavior and partial functionality
- **User Experience**: Errors can return meaningful HTTP status codes and messages

**Rule of Thumb**: If you're not in `main.go` and you're considering using `Must`, you probably shouldn't.