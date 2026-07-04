# Lesson 12: Middleware & Composition

## What We Built
```
go-shipit/
└── internal/
    └── transport/
        └── http/
            └── auth.go          ← Auth, rate limit, CORS, recover, structured logging, Chain()
```

---

## The Core Concept: Function Composition

In Go, middleware is just a function that wraps a handler:

```go
type Middleware = func(http.Handler) http.Handler
```

You **compose** them by wrapping one inside another:

```go
handler := Chain(mux,
    WithRecover,           // 1. Outermost — catches panics
    WithCORS("*"),         // 2. Sets CORS headers
    WithRequestID,         // 3. Adds request ID
    WithRateLimit(100),    // 4. Rate limiting
    WithStructuredLogging, // 5. Logs requests as JSON
    WithTimeout(10*s),     // 6. Innermost — sets deadline
)
```

Request flows: `Recover → CORS → RequestID → RateLimit → Logging → Timeout → Handler`

---

## Comparison: Middleware Registration

| Go | C# (ASP.NET) | Java (Spring) |
|-----|---------------|---------------|
| `Chain(mux, WithA, WithB)` | `app.UseA(); app.UseB();` | `@Order(1) FilterA`, `@Order(2) FilterB` |
| Function wrapping | Pipeline builder | Filter chain |
| Explicit order | Explicit order | Annotation order |

---

## Pattern 1: Authentication Middleware (Closure)

```go
func WithAuth(apiKey string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Header.Get("X-API-Key") != apiKey {
                writeError(w, 401, "invalid API key")
                return  // ← short-circuit: don't call next
            }
            next.ServeHTTP(w, r)  // ← proceed to next middleware/handler
        })
    }
}
```

### Key insight: The `apiKey` is "closed over" — captured by the closure at creation time.

### C# Equivalent:
```csharp
public class ApiKeyMiddleware {
    private readonly RequestDelegate _next;
    private readonly string _apiKey;

    public ApiKeyMiddleware(RequestDelegate next, IOptions<ApiKeyOptions> options) {
        _next = next;
        _apiKey = options.Value.Key;
    }

    public async Task InvokeAsync(HttpContext context) {
        if (context.Request.Headers["X-API-Key"] != _apiKey) {
            context.Response.StatusCode = 401;
            return;
        }
        await _next(context);
    }
}
```

### Java Equivalent:
```java
@Component
public class ApiKeyFilter extends OncePerRequestFilter {
    @Value("${api.key}") private String apiKey;

    @Override
    protected void doFilterInternal(HttpServletRequest req, HttpServletResponse res,
                                     FilterChain chain) throws IOException, ServletException {
        if (!apiKey.equals(req.getHeader("X-API-Key"))) {
            res.sendError(401, "invalid API key");
            return;
        }
        chain.doFilter(req, res);
    }
}
```

---

## Pattern 2: Rate Limiting (Shared State + Mutex)

```go
func WithRateLimit(rpm int) func(http.Handler) http.Handler {
    var mu sync.Mutex
    clients := make(map[string]*client)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            mu.Lock()
            // check/update rate limit for r.RemoteAddr
            mu.Unlock()
            next.ServeHTTP(w, r)
        })
    }
}
```

### Why sync.Mutex?
Each HTTP request runs in its own goroutine. The `clients` map is shared across all goroutines. Without a mutex, concurrent reads/writes = race condition → crash.

### C# Equivalent:
```csharp
// Built-in rate limiter in .NET 7+
builder.Services.AddRateLimiter(options => {
    options.AddFixedWindowLimiter("api", opts => {
        opts.PermitLimit = 100;
        opts.Window = TimeSpan.FromMinutes(1);
    });
});
app.UseRateLimiter();
```

### Java Equivalent:
```java
// Bucket4j or Guava RateLimiter
RateLimiter limiter = RateLimiter.create(100.0 / 60);  // 100/min
if (!limiter.tryAcquire()) {
    response.sendError(429);
    return;
}
```

---

## Pattern 3: Panic Recovery (defer + recover)

```go
func WithRecover(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v", err)
                writeError(w, 500, "internal server error")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### What is `recover()`?
- `panic()` = throw an unrecoverable error (like `throw` in C#/Java)
- `recover()` = catch it (only works inside `defer`)
- Without recovery middleware, one panic crashes the entire server

### C# Equivalent:
```csharp
app.UseExceptionHandler("/error");
// Catches unhandled exceptions, returns 500
```

### Java Equivalent:
```java
@ControllerAdvice
public class GlobalExceptionHandler {
    @ExceptionHandler(Exception.class)
    public ResponseEntity<String> handle(Exception ex) {
        return ResponseEntity.status(500).body("Internal Server Error");
    }
}
```

---

## Pattern 4: Structured Logging (JSON)

```go
entry := map[string]any{
    "level":       "info",
    "msg":         "http_request",
    "method":      r.Method,
    "path":        r.URL.Path,
    "status":      wrapped.status,
    "duration_ms": duration.Milliseconds(),
    "request_id":  reqID,
}
logJSON, _ := json.Marshal(entry)
log.Println(string(logJSON))
```

Output:
```json
{"level":"info","msg":"http_request","method":"GET","path":"/health","status":200,"duration_ms":1,"request_id":"20260704-170000.123"}
```

### Why structured (JSON) logs?
- Machine-parseable → searchable in Grafana, Datadog, Azure Monitor
- Each field is filterable: "show me all requests with status >= 500"
- In production, you'd use `slog` (Go 1.21+) or `zerolog` instead of this manual approach

### C# Equivalent (Serilog):
```csharp
Log.Information("HTTP {Method} {Path} responded {StatusCode} in {Duration}ms",
    method, path, statusCode, duration);
// Outputs JSON when configured with JsonFormatter
```

### Java Equivalent (Logback + Logstash):
```xml
<encoder class="net.logstash.logback.encoder.LogstashEncoder"/>
```
```java
log.info("http_request", kv("method", method), kv("path", path), kv("status", status));
```

---

## Pattern 5: The Chain Helper

```go
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}
```

This applies middleware in reverse so the **first in the list = outermost**:
```go
handler := Chain(mux, A, B, C)
// Execution order: A → B → C → handler → C → B → A
```

### C# Equivalent:
ASP.NET's pipeline builder does this implicitly:
```csharp
app.UseA();  // outermost
app.UseB();
app.UseC();  // innermost
```

---

## Composition in Go vs C#/Java

| Aspect | Go | C# | Java |
|--------|-----|-----|------|
| **Pattern** | Higher-order functions | Middleware pipeline | Filter chain |
| **DI for middleware** | Closures capture deps | Constructor injection | `@Autowired` |
| **Registration** | `Chain(mux, ...)` | `app.UseXxx()` | `@Bean FilterRegistration` |
| **Config** | Close over values: `WithAuth(key)` | `IOptions<T>` injection | `@Value` / `@ConfigurationProperties` |
| **Short-circuit** | `return` (don't call next) | Don't call `await next()` | Don't call `chain.doFilter()` |

---

## Try It

```bash
go build -o shipit.exe ./cmd/shipit
.\shipit.exe serve

# Test rate limiting:
for ($i=0; $i -lt 5; $i++) { (Invoke-WebRequest http://localhost:8080/health).StatusCode }

# Test CORS:
curl -I -X OPTIONS http://localhost:8080/deploys
# → Access-Control-Allow-Origin: *

# Structured log output on server:
# {"level":"info","msg":"http_request","method":"GET","path":"/health","status":200,"duration_ms":0,...}
```

---

## Key Takeaways

1. **Middleware = `func(http.Handler) http.Handler`** — wraps a handler with behavior.
2. **Closures capture config** — `WithAuth(key)` closes over the API key.
3. **`Chain()` helper** — readable middleware composition.
4. **`sync.Mutex`** — protects shared state across goroutines (rate limiter).
5. **`defer` + `recover()`** — catches panics so one bad request doesn't crash the server.
6. **Structured logging** — JSON logs for production observability.
7. **Short-circuit** — `return` without calling `next` to reject requests early.

---

## Next: [Lesson 13 — AI Integration](./13-ai-integration.md)
We'll build the risk scorer using GitHub Models (GPT-4o) for deploy risk assessment.
