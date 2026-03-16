# Standard Library First Examples

## ✅ Do: Standard library solutions

```go
// Use standard library for HTTP routing
mux := http.NewServeMux()
mux.HandleFunc("/users", handleUsers)
mux.HandleFunc("/health", handleHealth)

// Use standard library for JSON
json.Marshal(data)
json.Unmarshal(bytes, &result)

// Use standard library for time
time.Now()
time.Sleep(5 * time.Second)

// Use standard library for context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```


## ❌ Do Not: Unnecessary external dependencies

```go
// DON'T: Use gin/echo when standard library suffices for simple APIs
router := gin.Default()
router.GET("/users", handleUsers)

// DON'T: Use custom JSON library without good reason
import "github.com/json-iterator/go"

// DON'T: Use third-party time library for basic operations
import "github.com/jinzhu/now"
```
