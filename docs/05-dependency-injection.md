# Lesson 05: Dependency Injection (No Framework)

## What We Built
```
go-shipit/
├── cmd/shipit/main.go                    ← Composition root (wires deps)
├── internal/
│   ├── domain/
│   │   └── service.go                    ← DeployService (depends on interfaces)
│   └── adapters/
│       └── inmemory/
│           ├── repository.go             ← In-memory DeployRepository
│           ├── queue.go                  ← In-memory QueuePublisher
│           └── notifier.go              ← In-memory Notifier
```

---

## The Core Difference

| | Go | C# | Java |
|-|-----|-----|------|
| **DI mechanism** | Manual (constructor functions) | Framework (`IServiceCollection`) | Framework (Spring `@Autowired`) |
| **Registration** | None — you call `New()` yourself | `services.AddScoped<IFoo, Foo>()` | `@Component` / `@Bean` |
| **Resolution** | None — you pass the dependency | Container resolves at runtime | Container resolves at runtime |
| **Magic** | Zero | Moderate (reflection, lifetime scopes) | Heavy (classpath scanning, proxies) |

---

## How DI Works in Go

### Step 1: Service depends on interfaces (not concrete types)

```go
type DeployService struct {
    repo     DeployRepository  // interface — could be MySQL, Postgres, or in-memory
    queue    QueuePublisher    // interface — could be Azure Queue, RabbitMQ, or stdout
    notifier Notifier          // interface — could be Slack, email, or a mock
}
```

### Step 2: Constructor accepts interfaces

```go
func NewDeployService(repo DeployRepository, queue QueuePublisher, notifier Notifier) *DeployService {
    return &DeployService{repo: repo, queue: queue, notifier: notifier}
}
```

### Step 3: main() wires everything (the "composition root")

```go
func main() {
    // Create concrete implementations
    repo := inmemory.NewRepository()
    queue := inmemory.NewQueue()
    notifier := inmemory.NewNotifier()

    // Inject into service
    service := domain.NewDeployService(repo, queue, notifier)

    // Use service — it has no idea what's behind the interfaces
    service.StartDeploy(...)
}
```

That's it. No framework. No container. No reflection. No magic.

---

## C# Equivalent

```csharp
// Program.cs (or Startup.cs)
var builder = WebApplication.CreateBuilder(args);

// Register services — the container manages lifetimes
builder.Services.AddScoped<IDeployRepository, MySqlDeployRepository>();
builder.Services.AddScoped<IQueuePublisher, AzureQueuePublisher>();
builder.Services.AddScoped<INotifier, SlackNotifier>();
builder.Services.AddScoped<DeployService>();

var app = builder.Build();

// The container resolves DeployService and all its dependencies automatically
app.MapPost("/deploy", (DeployService svc) => svc.StartDeploy(...));
```

### Differences from Go:
- **Lifetime management** (Scoped/Singleton/Transient) — Go doesn't have this; you manage lifetimes yourself
- **Automatic resolution** — C# container figures out the dependency graph
- **Runtime errors** — If you forget to register something, you get a runtime exception. In Go, you get a **compile error** (missing argument).

---

## Java (Spring) Equivalent

```java
@Service
public class DeployService {
    private final DeployRepository repo;
    private final QueuePublisher queue;
    private final Notifier notifier;

    @Autowired  // Spring resolves these from the application context
    public DeployService(DeployRepository repo, QueuePublisher queue, Notifier notifier) {
        this.repo = repo;
        this.queue = queue;
        this.notifier = notifier;
    }
}

@Repository
public class MySqlDeployRepository implements DeployRepository { ... }

@Component
public class SlackNotifier implements Notifier { ... }
```

### Differences from Go:
- **Classpath scanning** — Spring finds implementations automatically via annotations
- **Proxy magic** — Spring creates proxies for transactions, AOP, etc.
- **Circular dependencies** — Spring can resolve them (sometimes). Go can't — and that's a feature.

---

## Why Go's Approach Is Better (for this use case)

| Advantage | Explanation |
|-----------|-------------|
| **Compile-time safety** | Forget a dependency → compile error. Not a runtime crash. |
| **No hidden magic** | Read main() and you see every dependency clearly |
| **Easy to trace** | Ctrl+Click through the code — no framework interception |
| **Fast startup** | No classpath scanning, no reflection, no proxy generation |
| **Testable** | Swap any dependency in tests — just pass a different struct |

### The tradeoff:
In a large app (100+ services), manual wiring gets verbose. Some Go teams use [Wire](https://github.com/google/wire) (compile-time DI code generator) to auto-generate the wiring. But for most projects, manual wiring is fine.

---

## The Adapter Pattern in Action

Our `inmemory` adapters satisfy the domain interfaces **without importing them**:

```go
// internal/adapters/inmemory/repository.go
type Repository struct {
    deployments map[string]*domain.Deployment
}

func (r *Repository) GetByID(id string) (*domain.Deployment, error) { ... }
func (r *Repository) Save(d *domain.Deployment) error { ... }
func (r *Repository) ListByService(name string, limit int) ([]*domain.Deployment, error) { ... }
```

This struct satisfies `domain.DeployRepository` because it has the right methods. Later, we'll create a `mysql.Repository` with the same methods — and swap it in with zero changes to the domain.

---

## Swapping Implementations

```go
// Local development
repo := inmemory.NewRepository()

// Production (future lesson)
repo := mysql.NewRepository(db)

// Testing
repo := &MockRepository{...}

// Same service, different adapters:
service := domain.NewDeployService(repo, queue, notifier)
```

---

## Try It

```bash
go build -o shipit.exe ./cmd/shipit
.\shipit.exe demo
```

Output:
```
--- Dependency Injection (Lesson 05) ---
  [queue] → deployments: {"deployment_id":"deploy-001"}
  [notify] #deploys: 🚀 New deploy: payments-api v2.4.1 → production
✓ Created deployment: deploy-001 (pending)
✓ Fetched from repo: payments-api → production

--- Error Handling (Lesson 04) ---
✓ Expected error: GetDeploy(does-not-exist): deployment "does-not-exist": not found
```

---

## Key Takeaways

1. **DI in Go = pass interfaces to constructors.** No framework needed.
2. **main() is the composition root** — where all dependencies are wired.
3. **Compile-time safety** — forget a dep, get a compile error (not a runtime crash).
4. **No magic** — `Ctrl+Click` works. No proxies. No reflection.
5. **Swap adapters freely** — inmemory for dev, MySQL for prod, mocks for tests.
6. **Interfaces defined at the consumer** — the domain defines what it needs.

---

## Next: [Lesson 06 — HTTP Server](./06-http-server.md)
We'll build the httpservice using Go's standard `net/http` package — no framework (no Gin, no Echo).
