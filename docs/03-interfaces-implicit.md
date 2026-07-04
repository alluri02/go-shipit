# Lesson 03: Interfaces (Implicit Implementation)

## What We Built
```
go-shipit/
└── internal/
    └── ports/
        └── ports.go    ← 6 interface definitions (the "ports" in hexagonal architecture)
```

---

## The Big Idea

**Go interfaces are implemented implicitly.** There is no `implements` keyword.

If your struct has the right methods → it satisfies the interface. Period.

```go
// Define the interface
type Notifier interface {
    Notify(channel, message string) error
}

// This struct implements Notifier — WITHOUT saying so anywhere
type SlackNotifier struct {
    webhookURL string
}

func (s *SlackNotifier) Notify(channel, message string) error {
    // send to Slack...
    return nil
}
```

The compiler checks at the **point of use**, not at the point of definition:

```go
func alertTeam(n Notifier) {      // ← accepts any Notifier
    n.Notify("#deploys", "done!")
}

slack := &SlackNotifier{webhookURL: "https://..."}
alertTeam(slack)  // ✓ Compiles — SlackNotifier has Notify() method
```

---

## Concept Map: Go vs C# vs Java

| Aspect | Go | C# | Java |
|--------|-----|-----|------|
| **Declaration** | `type X interface { ... }` | `public interface IX { ... }` | `public interface X { ... }` |
| **Implementation** | Implicit (just have the methods) | Explicit: `class Foo : IX` | Explicit: `class Foo implements X` |
| **Naming convention** | `Notifier`, `Reader`, `Deployer` | `INotifier`, `IReader`, `IDeployer` | `Notifier`, `Reader`, `Deployer` |
| **Multiple interfaces** | Automatic (satisfy all you want) | `class Foo : IA, IB` | `class Foo implements A, B` |
| **Empty interface** | `interface{}` or `any` | `object` | `Object` |
| **Where defined** | Near the **consumer** (who needs it) | Near the **producer** (who implements it) | Near the **producer** |
| **Default methods** | Not possible | `default` in C# 8+ | `default` in Java 8+ |

---

## Deep Dive: Why Implicit Interfaces Matter

### 1. You Can Satisfy Interfaces You've Never Seen

```go
// Some library defines:
type Writer interface {
    Write(p []byte) (n int, err error)
}

// Your struct implements it WITHOUT importing the library:
type FileLogger struct{}

func (f *FileLogger) Write(p []byte) (int, error) {
    // write to file...
    return len(p), nil
}
```

In C#/Java, you'd need to explicitly reference the interface:
```csharp
// C# — must declare the relationship
class FileLogger : IWriter { ... }
```
```java
// Java — must declare the relationship
class FileLogger implements Writer { ... }
```

### 2. Decoupling Is Free

In Go, the **consumer** defines what it needs. The **producer** doesn't even need to know the interface exists.

```go
// In ports/ports.go — the domain defines what it needs:
type DeployRepository interface {
    GetByID(id string) (*Deployment, error)
    Save(deployment *Deployment) error
}

// In adapters/mysql/repository.go — the adapter just has the methods:
type MySQLRepository struct { db *sql.DB }

func (r *MySQLRepository) GetByID(id string) (*Deployment, error) { ... }
func (r *MySQLRepository) Save(d *Deployment) error { ... }

// MySQLRepository satisfies DeployRepository — no coupling between the two files
```

### 3. Easy Mocking for Tests

```go
// In tests, create a mock that satisfies the interface:
type MockRepository struct {
    deployments map[string]*Deployment
}

func (m *MockRepository) GetByID(id string) (*Deployment, error) {
    d, ok := m.deployments[id]
    if !ok {
        return nil, errors.New("not found")
    }
    return d, nil
}

func (m *MockRepository) Save(d *Deployment) error {
    m.deployments[d.ID] = d
    return nil
}
```

No mocking framework needed. No `Moq`, no `Mockito`. Just a struct with the right methods.

---

## Deep Dive: Interface Design Best Practices in Go

### Keep Interfaces Small (1-3 methods)

Go's philosophy: **"The bigger the interface, the weaker the abstraction."** — Rob Pike

```go
// ✓ Good — small, focused
type Reader interface {
    Read(p []byte) (n int, err error)
}

// ✗ Bad — too many methods, hard to implement
type Repository interface {
    GetByID(id string) (*Thing, error)
    Save(t *Thing) error
    Delete(id string) error
    List(filter Filter) ([]*Thing, error)
    Count() (int, error)
    Search(query string) ([]*Thing, error)
    // ... 10 more methods
}
```

### C# Comparison: Interface Segregation

In C#, you'd use Interface Segregation Principle (ISP) to split large interfaces:
```csharp
public interface IReadRepository<T> {
    Task<T> GetByIdAsync(string id);
    Task<IEnumerable<T>> ListAsync();
}

public interface IWriteRepository<T> {
    Task SaveAsync(T entity);
    Task DeleteAsync(string id);
}
```

Go naturally encourages this because implicit implementation makes small interfaces free.

### Java Comparison

Java's Spring framework often has large repository interfaces:
```java
public interface JpaRepository<T, ID> extends PagingAndSortingRepository<T, ID> {
    // 20+ methods
}
```

Go's standard library avoids this. `io.Reader` has 1 method. `io.Writer` has 1 method. Compose them when needed:
```go
type ReadWriter interface {
    Reader
    Writer
}
```

---

## Deep Dive: Interface Composition (Embedding)

Go interfaces can embed other interfaces — like interface inheritance:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Composed interface — requires BOTH Read and Write
type ReadWriter interface {
    Reader
    Writer
}
```

### C# equivalent:
```csharp
public interface IReadWriter : IReader, IWriter { }
```

### Java equivalent:
```java
public interface ReadWriter extends Reader, Writer { }
```

---

## Deep Dive: Compile-Time Interface Checks

Go checks interfaces at compile time, but only at the point of use. To force an early check:

```go
// Verify at compile time that MySQLRepository implements DeployRepository
var _ ports.DeployRepository = (*MySQLRepository)(nil)
```

This is a common Go idiom. It creates a zero-value variable and assigns it to the interface type. If the methods don't match, you get a compile error immediately.

### C#/Java don't need this because they check at the declaration:
```csharp
class MySQLRepository : IDeployRepository { }  // Compile error if methods missing
```

---

## Our Interfaces (ShipIt Ports)

| Interface | Methods | Purpose |
|-----------|---------|---------|
| `DeployRepository` | `GetByID`, `Save`, `ListByService` | Persist deployments (MySQL) |
| `QueuePublisher` | `Publish` | Enqueue deploy jobs (Azure Queue) |
| `QueueConsumer` | `Consume` | Read deploy jobs (Azure Queue) |
| `ImageBuilder` | `Build`, `Push` | Build & push containers (ACR) |
| `Deployer` | `Deploy`, `Rollback` | Deploy to target (Container Apps) |
| `Notifier` | `Notify` | Send notifications (Slack) |
| `RiskAnalyzer` | `Assess` | AI risk scoring (GitHub Models) |

Each will get a real adapter in later lessons and a mock for testing.

---

## Try It

The interfaces don't "do" anything yet — they're contracts. But you can verify they compile:

```bash
go build ./...
```

In the next lesson, we'll define domain errors and see how Go handles errors vs exceptions.

---

## Key Takeaways

1. **No `implements` keyword** — just have the right methods.
2. **Consumer defines the interface** — not the producer. This inverts the dependency.
3. **Keep interfaces small** — 1-3 methods is ideal in Go.
4. **Composition over inheritance** — embed interfaces to build larger ones.
5. **Mocking is free** — any struct with the right methods is a valid mock.
6. **Compile-time check idiom** — `var _ Interface = (*Type)(nil)`

---

## Next: [Lesson 04 — Error Handling](./04-error-handling.md)
We'll define domain-specific errors and learn Go's explicit error handling vs exceptions.
