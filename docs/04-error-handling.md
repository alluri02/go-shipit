# Lesson 04: Error Handling

## What We Built
```
go-shipit/
└── internal/
    └── domain/
        ├── errors.go       ← Sentinel errors + custom error types
        └── validate.go     ← Validation with error returns
```

---

## The Fundamental Difference

| | Go | C# | Java |
|-|-----|-----|------|
| **Mechanism** | Return values | Exceptions (throw/catch) | Exceptions (throw/catch) |
| **Control flow** | Explicit `if err != nil` | Hidden (try/catch anywhere up the stack) | Hidden (try/catch anywhere up the stack) |
| **Performance** | Zero-cost (just a return) | Expensive (stack trace capture) | Expensive (stack trace capture) |
| **Visibility** | Always visible in the code | Can be invisible (uncaught) | Can be invisible (unchecked) |

---

## Pattern 1: Sentinel Errors (Predefined Error Values)

```go
// Define once at package level
var ErrNotFound = errors.New("not found")

// Return it
func (r *Repo) GetByID(id string) (*Deployment, error) {
    d, ok := r.store[id]
    if !ok {
        return nil, ErrNotFound
    }
    return d, nil
}

// Check it
deploy, err := repo.GetByID("xyz")
if errors.Is(err, domain.ErrNotFound) {
    // handle 404
}
```

### C# Equivalent
```csharp
// Throw a typed exception
throw new NotFoundException($"Deployment {id} not found");

// Catch it
try {
    var deploy = await repo.GetByIdAsync(id);
} catch (NotFoundException ex) {
    // handle 404
}
```

### Java Equivalent
```java
// Throw
throw new NotFoundException("Deployment " + id + " not found");

// Catch
try {
    Deployment deploy = repo.findById(id);
} catch (NotFoundException ex) {
    // handle 404
}
```

---

## Pattern 2: Custom Error Types (Structs That Implement `error`)

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s — %s", e.Field, e.Message)
}

// Usage: return an error
func ValidateDeployment(name string) error {
    if name == "" {
        return &ValidationError{Field: "name", Message: "required"}
    }
    return nil
}

// Usage: check for specific error type
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Field %s: %s\n", valErr.Field, valErr.Message)
}
```

### C# Equivalent
```csharp
public class ValidationException : Exception {
    public string Field { get; }
    public ValidationException(string field, string msg) : base(msg) {
        Field = field;
    }
}

// Catch and inspect
catch (ValidationException ex) {
    Console.WriteLine($"Field {ex.Field}: {ex.Message}");
}
```

### Java Equivalent
```java
public class ValidationException extends RuntimeException {
    private final String field;
    public ValidationException(String field, String msg) {
        super(msg);
        this.field = field;
    }
    public String getField() { return field; }
}
```

---

## Pattern 3: Wrapping Errors (Adding Context)

The `%w` verb in `fmt.Errorf` wraps an error, preserving the chain:

```go
// Wrap with context
func (s *DeployService) StartDeploy(id string) error {
    deploy, err := s.repo.GetByID(id)
    if err != nil {
        return fmt.Errorf("StartDeploy(%s): %w", id, err)  // wraps err
    }
    // ...
}

// The chain is preserved:
err := service.StartDeploy("xyz")
errors.Is(err, ErrNotFound)  // true! %w preserves the chain
```

### C# Equivalent: Inner Exceptions
```csharp
try {
    var deploy = await repo.GetByIdAsync(id);
} catch (Exception ex) {
    throw new DeployException($"StartDeploy({id}) failed", ex);  // inner exception
}
```

### Java Equivalent: Exception Chaining
```java
try {
    Deployment deploy = repo.findById(id);
} catch (Exception ex) {
    throw new DeployException("StartDeploy(" + id + ") failed", ex);  // cause
}
```

---

## Pattern 4: errors.Is vs errors.As

| Function | Purpose | C# Equivalent | Java Equivalent |
|----------|---------|---------------|-----------------|
| `errors.Is(err, target)` | Check if err wraps target (value comparison) | `ex is NotFoundException` | `ex instanceof NotFoundException` |
| `errors.As(err, &target)` | Extract specific error type from chain | `ex as ValidationException` | Cast after `instanceof` |

```go
// errors.Is — "is this error (or any wrapped error) equal to ErrNotFound?"
if errors.Is(err, ErrNotFound) {
    // 404
}

// errors.As — "extract the ValidationError from the chain"
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println(valErr.Field)  // access typed fields
}
```

---

## Pattern 5: The `if err != nil` Pattern

This is the most common code you'll write in Go:

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("context: %w", err)
}
// use result
```

### Why no try/catch?

Go designers believe:
1. **Error paths should be visible** — not hidden in catch blocks 3 levels up
2. **You should handle errors where they occur** — not let them bubble silently
3. **Performance matters** — exceptions capture stack traces (expensive)

### Common objection: "It's verbose!"

Yes. But you always know exactly what can fail and how it's handled. Compare:

```go
// Go — every error point is visible
user, err := repo.GetUser(id)
if err != nil {
    return nil, fmt.Errorf("get user: %w", err)
}
orders, err := repo.GetOrders(user.ID)
if err != nil {
    return nil, fmt.Errorf("get orders: %w", err)
}
```

```csharp
// C# — where can this fail? Who catches it? 🤷
var user = await repo.GetUserAsync(id);
var orders = await repo.GetOrdersAsync(user.Id);
// If either throws, it could be caught 5 levels up... or not at all
```

---

## ShipIt Error Handling Strategy

| Layer | Error Approach |
|-------|---------------|
| **Domain** | Define sentinel errors + custom types |
| **Adapters** | Wrap external errors with context: `fmt.Errorf("mysql: %w", err)` |
| **Transport (HTTP)** | Map domain errors to HTTP status codes |
| **Transport (Slack)** | Map domain errors to user-friendly messages |

```go
// In the HTTP handler (future lesson):
func handleGetDeploy(w http.ResponseWriter, r *http.Request) {
    deploy, err := service.GetDeploy(id)
    if errors.Is(err, domain.ErrNotFound) {
        http.Error(w, "deployment not found", http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // return deploy as JSON
}
```

---

## Try It

```bash
go build ./...
.\shipit.exe demo
```

The `demo` command shows validation and state transitions in action.

---

## Key Takeaways

1. **Errors are values** — returned, not thrown. Zero cost.
2. **`if err != nil`** — the most common pattern. Embrace it.
3. **Sentinel errors** (`ErrNotFound`) — for simple "what went wrong" checks.
4. **Custom error types** (`ValidationError`) — when you need structured context.
5. **`%w` wrapping** — adds context while preserving the error chain.
6. **`errors.Is` / `errors.As`** — traverse the chain to find specific errors.
7. **No hidden control flow** — you always see where errors are handled.

---

## Next: [Lesson 05 — Dependency Injection](./05-dependency-injection.md)
We'll wire the domain services together using constructor injection — no framework needed.
