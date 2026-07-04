# Lesson 11: Context & Cancellation

## What We Built
```
go-shipit/
└── internal/
    └── transport/
        └── http/
            ├── graceful.go      ← Graceful shutdown with OS signals
            └── middleware.go    ← Request timeout, logging, request ID
```

---

## The Core Difference

| | Go (`context.Context`) | C# (`CancellationToken`) | Java |
|-|------------------------|--------------------------|------|
| **Mechanism** | `context.Context` (first param) | `CancellationToken` (last param) | No direct equivalent |
| **Timeout** | `context.WithTimeout(ctx, 10s)` | `new CancellationTokenSource(10s)` | `CompletableFuture.orTimeout()` |
| **Cancellation** | `cancel()` function | `cts.Cancel()` | `future.cancel()` |
| **Propagation** | Pass as first arg everywhere | Pass as last arg everywhere | Thread interruption / Reactor |
| **Values** | `context.WithValue(ctx, key, val)` | `HttpContext.Items["key"]` | `ThreadLocal<T>` / MDC |

---

## What Is context.Context?

`context.Context` is an interface that carries:
1. **Deadlines** — "this operation must complete by X time"
2. **Cancellation** — "stop working, the caller gave up"
3. **Values** — request-scoped data (request ID, user ID, trace ID)

```go
// The convention: context is ALWAYS the first parameter
func (r *Repository) GetByID(ctx context.Context, id string) (*Deployment, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ...", id)
    // If ctx is cancelled (timeout/caller disconnect), the query aborts
}
```

---

## Pattern 1: context.WithTimeout

```go
// "This operation must complete within 5 seconds"
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()  // Always call cancel to release resources

result, err := slowOperation(ctx)
if err == context.DeadlineExceeded {
    // Operation took too long — timed out
}
```

### C# Equivalent:
```csharp
using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(5));
try {
    var result = await SlowOperationAsync(cts.Token);
} catch (OperationCanceledException) {
    // Timed out
}
```

### Java Equivalent:
```java
CompletableFuture<Result> future = CompletableFuture.supplyAsync(() -> slowOperation());
try {
    Result result = future.get(5, TimeUnit.SECONDS);
} catch (TimeoutException e) {
    // Timed out
}
```

---

## Pattern 2: context.WithCancel

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    // Do work with ctx...
    result, err := doWork(ctx)
}()

// Later: cancel the work
cancel()  // All operations using ctx will receive cancellation
```

### C# Equivalent:
```csharp
var cts = new CancellationTokenSource();
_ = Task.Run(() => DoWorkAsync(cts.Token));

// Later
cts.Cancel();
```

---

## Pattern 3: context.WithValue (Request-Scoped Data)

```go
// Store a value in context
ctx := context.WithValue(r.Context(), requestIDKey, "abc-123")

// Retrieve it downstream
requestID := ctx.Value(requestIDKey).(string)
```

### Rules for context values:
1. Keys must be unexported types (prevents collisions)
2. Only for request-scoped data (request ID, user, trace)
3. NEVER for optional function parameters (use struct/options pattern)

### C# Equivalent:
```csharp
// HttpContext.Items — request-scoped dictionary
ctx.Items["RequestId"] = "abc-123";
var id = (string)ctx.Items["RequestId"];
```

### Java Equivalent:
```java
// SLF4J MDC for logging context
MDC.put("requestId", "abc-123");
String id = MDC.get("requestId");
```

---

## Pattern 4: Graceful Shutdown

```go
// Listen for OS signals (Ctrl+C, docker stop, k8s terminate)
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// Block until signal received
<-quit

// Shutdown with timeout — waits for in-flight requests
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

### What happens during graceful shutdown:
1. Server stops accepting NEW connections
2. In-flight requests continue until they complete OR timeout
3. After all requests finish (or timeout), server exits

### C# Equivalent:
```csharp
// ASP.NET Core handles this automatically with IHostLifetime
var app = builder.Build();
await app.RunAsync();  // Handles SIGTERM, waits for requests
// Or manually:
appLifetime.ApplicationStopping.Register(() => {
    // cleanup code
});
```

### Java Equivalent:
```java
// Spring Boot handles this with:
server.shutdown=graceful
spring.lifecycle.timeout-per-shutdown-phase=30s

// Or manual:
Runtime.getRuntime().addShutdownHook(new Thread(() -> {
    server.shutdown();
    server.awaitTermination(30, TimeUnit.SECONDS);
}));
```

---

## Pattern 5: Middleware Chain

Middleware wraps handlers — each adds behavior before/after:

```go
var handler http.Handler = mux           // Base router
handler = WithTimeout(10*time.Second)(handler)  // Add timeout
handler = WithLogging(handler)                   // Add logging
handler = WithRequestID(handler)                 // Add request ID
```

Request flows through: `RequestID → Logging → Timeout → Router → Handler`

### C# Equivalent:
```csharp
app.UseMiddleware<RequestIdMiddleware>();
app.UseHttpLogging();
app.UseRouting();
app.MapControllers();
// Order matters! First registered = outermost
```

### Java Equivalent:
```java
@Bean
public FilterRegistrationBean<RequestIdFilter> requestIdFilter() { ... }

@Bean
public FilterRegistrationBean<LoggingFilter> loggingFilter() { ... }
// @Order controls execution sequence
```

---

## Deep Dive: The select Statement

```go
select {
case sig := <-quit:       // OS signal received
    log.Printf("Signal: %v", sig)
case err := <-errCh:      // Server error
    return err
case <-ctx.Done():        // Context cancelled/timed out
    return ctx.Err()
}
```

`select` blocks until ONE of the cases is ready. It's like a switch for channels.

### C# Equivalent:
```csharp
// Task.WhenAny — wait for first to complete
var completed = await Task.WhenAny(signalTask, serverTask, cancellationTask);
```

### Java Equivalent:
```java
// CompletableFuture.anyOf
CompletableFuture.anyOf(signalFuture, serverFuture).get();
```

---

## Try It

```bash
go build -o shipit.exe ./cmd/shipit
.\shipit.exe serve

# In another terminal, make a request:
curl http://localhost:8080/health
# Notice the log output: [20260704-163012.000] GET /health 200 1.2ms

# Press Ctrl+C — watch graceful shutdown:
# "Received signal: interrupt. Shutting down gracefully..."
# "Server stopped gracefully"
```

---

## Key Takeaways

1. **`context.Context`** = first parameter everywhere. Carries timeouts + cancellation + values.
2. **`context.WithTimeout`** — auto-cancels after duration. Use for every I/O operation.
3. **`defer cancel()`** — ALWAYS call cancel. Prevents resource leaks.
4. **Graceful shutdown** — catch SIGINT/SIGTERM, call `server.Shutdown(ctx)`.
5. **`select`** — wait on multiple channels simultaneously.
6. **Middleware** — `func(http.Handler) http.Handler`. Chain them around your mux.
7. **Context values** — for request-scoped data only (request ID, user, trace).

---

## Next: [Lesson 12 — Middleware & Composition](./12-middleware.md)
We'll add authentication, rate limiting, and structured logging middleware.
