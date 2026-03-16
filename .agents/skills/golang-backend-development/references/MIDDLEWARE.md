# Middleware Examples

## ✅ Correct: Middleware implementation

```go
// internal/middleware/logging.go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(wrapped, r)
        
        log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
    })
}

// internal/middleware/auth.go
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "missing authorization header", http.StatusUnauthorized)
                return
            }
            
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            userID, err := validateJWT(tokenString, jwtSecret)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }
            
            // Add userID to context
            ctx := context.WithValue(r.Context(), "userID", userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// internal/middleware/cors.go
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Usage in main.go or router setup
mux := http.NewServeMux()
mux.Handle("/api/users", handler.NewUserHandler(svc))

handler := CORSMiddleware(LoggingMiddleware(AuthMiddleware(jwtSecret)(mux)))
http.ListenAndServe(":8080", handler)
```
