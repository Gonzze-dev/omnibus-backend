# The Handler Wrapper Pattern

## Overview
The `Handler` wrapper is a generic HTTP handler implementation that standardizes the request/response lifecycle. It decouples the HTTP transport mechanics (decoding, encoding, status codes) from the business logic.

## Base Implementation (`internal/handler/base.go`)

```go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// Handler wraps a strongly-typed logic function with HTTP transport handling.
type Handler[REQ any, RESP any] struct {
	// HandleFunc is the core business logic.
	HandleFunc func(ctx context.Context, req REQ) (RESP, error)

	// PreHandlers are executed after decoding but before the main handler.
	// Useful for validation or other checks.
	PreHandlers []func(ctx context.Context, req REQ) error

	// PostHandlers are executed after the main handler but before encoding.
	// Useful for auditing, additional logging, etc.
	PostHandlers []func(ctx context.Context, req REQ, resp RESP) error

	// DecodeFunc transforms the raw HTTP request into the typed REQ struct.
	// Defaults to DefaultDecoder if nil.
	DecodeFunc func(r *http.Request) (REQ, error)

	// EncodeFunc transforms the typed RESP struct into the HTTP response.
	// Defaults to DefaultEncoder if nil.
	EncodeFunc func(w http.ResponseWriter, resp RESP) error

	// ErrorHandler maps errors to HTTP status codes and responses.
	// Defaults to DefaultErrorHandler if nil.
	ErrorHandler func(w http.ResponseWriter, err error)
}

// ServeHTTP implements http.Handler.
func (h Handler[REQ, RESP]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Setup Defaults
	decode := h.DecodeFunc
	if decode == nil {
		decode = DefaultDecoder[REQ]
	}
	encode := h.EncodeFunc
	if encode == nil {
		encode = DefaultEncoder[RESP]
	}
	handleError := h.ErrorHandler
	if handleError == nil {
		handleError = DefaultErrorHandler
	}

	// 2. Decode
	req, err := decode(r)
	if err != nil {
		handleError(w, err)
		return
	}

	ctx := r.Context()

	// 3. Pre-handling
	for _, pre := range h.PreHandlers {
		if err := pre(ctx, req); err != nil {
			handleError(w, err)
			return
		}
	}

	// 4. Handle Logic
	resp, err := h.HandleFunc(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	// 5. Post-handling
	for _, post := range h.PostHandlers {
		if err := post(ctx, req, resp); err != nil {
			handleError(w, err)
			return
		}
	}

	// 6. Encode Response
	if err := encode(w, resp); err != nil {
		handleError(w, err)
	}
}

// DefaultDecoder parses JSON body into the target struct.
func DefaultDecoder[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, err // In a real app, wrap this error to indicate 400 Bad Request
	}
	return v, nil
}

// DefaultEncoder writes the response struct as JSON.
func DefaultEncoder[T any](w http.ResponseWriter, v T) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// DefaultErrorHandler handles errors.
func DefaultErrorHandler(w http.ResponseWriter, err error) {
	// In a real app, inspect 'err' type to determine 400 vs 500
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
```

## Usage Example (`internal/handler/user.go`)

```go
package handler

import (
	"context"
	"net/http"

	"your-project/internal/models"
	"your-project/internal/service"
)

// NewCreateUserHandler creates an HTTP handler for user creation.
func NewCreateUserHandler(svc service.UserService) http.Handler {
	return Handler[models.CreateUserRequest, models.UserResponse]{
		HandleFunc: func(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
			// Call the service layer
			user, err := svc.CreateUser(ctx, req)
			if err != nil {
				return models.UserResponse{}, err
			}
			// Map domain model to DTO
			return models.UserResponse{
				ID:    user.ID,
				Email: user.Email,
			}, nil
		},
		// Optional: Override defaults if needed
		// DecodeFunc: CustomDecoder,
	}
}
```

## Benefits
- **Consistency**: Every endpoint behaves the same way regarding JSON parsing and error formatting.
- **Type Safety**: The compiler ensures the request and response types match what the handler expects.
- **Testability**: The `HandleFunc` logic is pure and easy to test in isolation if extracted.
- **Separation of Concerns**: HTTP details are hidden in `ServeHTTP`, while business logic stays in `HandleFunc`.

## Flexibility (Note)

Depending on requirements, the function signatures for `PreHandlers`, `PostHandlers`, `DecodeFunc`, `EncodeFunc`, and `ErrorHandler` can be adjusted and do not strictly need to follow the example.

For instance:
1. `PreHandlers` can be changed to `func(ctx context.Context, req *REQ) context.Context`
2. `PostHandlers` can be changed to `func(ctx context.Context, w http.ResponseWriter, resp RESP) error`
3. `DecodeFunc` can be changed to `func(ctx context.Context, r *http.Request) (*REQ, error)`
4. `EncodeFunc` can be changed to `func(ctx context.Context, w http.ResponseWriter, resp RESP) error`
