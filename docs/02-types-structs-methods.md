# Lesson 02: Types, Structs & Methods

## What We Built
```
internal/domain/
├── version.go        ← (from lesson 01)
├── status.go         ← Custom type + iota (Go's "enum")
├── environment.go    ← Struct + factory function + methods
└── deployment.go     ← Main model: pointer receivers, composition
```

---

## Concept Map: Go vs C# vs Java

| Concept | Go | C# | Java |
|---------|-----|-----|------|
| **Class** | `type X struct {}` | `class X {}` | `class X {}` |
| **Enum** | `type X int` + `iota` | `enum X {}` | `enum X {}` |
| **Constructor** | `func NewX() *X` (convention) | `public X()` | `public X()` |
| **Method** | `func (x *X) DoThing()` | `public void DoThing()` | `public void doThing()` |
| **this/self** | Named receiver: `(d *Deployment)` | Implicit `this` | Implicit `this` |
| **Getter** | Just a method: `func (x X) Name() string` | `public string Name { get; }` | `public String getName()` |
| **Pass by ref** | Pointer receiver `*X` | `class` (always ref) | Always ref for objects |
| **Pass by value** | Value receiver `X` | `struct` | Primitives only |
| **Inheritance** | ❌ None | `class Child : Parent` | `class Child extends Parent` |
| **Composition** | Embed struct in struct | Has-a field | Has-a field |
| **ToString** | `func (x X) String() string` | `override string ToString()` | `@Override String toString()` |

---

## Deep Dive: Structs (Go's "Classes")

Go has **no classes**. It has structs + methods. This is a deliberate design choice.

```go
// Go
type Deployment struct {
    ID          string
    ServiceName string
    Status      DeployStatus
    CreatedAt   time.Time
}
```

```csharp
// C#
public class Deployment
{
    public string Id { get; set; }
    public string ServiceName { get; set; }
    public DeployStatus Status { get; set; }
    public DateTime CreatedAt { get; set; }
}
```

```java
// Java
public class Deployment {
    private String id;
    private String serviceName;
    private DeployStatus status;
    private LocalDateTime createdAt;

    // + getters, setters, constructor...
}
```

### Key Differences:
1. **No access modifiers on fields** — uppercase = public, lowercase = private
2. **No getters/setters** — fields are accessed directly (Go philosophy: less ceremony)
3. **No `this` keyword** — the receiver is explicitly named: `(d *Deployment)`
4. **Zero value** — every struct field has a default (empty string, 0, false, zero time)

---

## Deep Dive: Enums via `iota`

Go doesn't have a built-in `enum` keyword. Instead:

```go
type DeployStatus int

const (
    DeployStatusPending   DeployStatus = iota // 0
    DeployStatusBuilding                      // 1
    DeployStatusPushing                       // 2
    DeployStatusDeploying                     // 3
)
```

`iota` starts at 0 and auto-increments within a `const` block.

### C# Equivalent
```csharp
public enum DeployStatus
{
    Pending = 0,
    Building = 1,
    Pushing = 2,
    Deploying = 3
}
```

### Java Equivalent
```java
public enum DeployStatus {
    PENDING, BUILDING, PUSHING, DEPLOYING
}
```

### Why no real enums?
Go's philosophy: keep the language small. Custom types + `iota` give you type safety without adding a language feature. The trade-off: no exhaustive switch checking at compile time (you handle `default:` manually).

---

## Deep Dive: Methods & Receivers

In Go, methods are just functions with a **receiver** — the struct they operate on.

### Value Receiver (reads only, works on a copy)
```go
func (d Deployment) IsHighRisk() bool {
    return d.RiskScore >= 7
}
```

### Pointer Receiver (can modify the original)
```go
func (d *Deployment) Advance(next DeployStatus) {
    d.Status = next       // modifies the actual struct
    d.UpdatedAt = time.Now()
}
```

### When to use which?

| Use | When |
|-----|------|
| Value receiver `(x X)` | Method only reads data, struct is small |
| Pointer receiver `(x *X)` | Method modifies data, OR struct is large (avoids copy) |

### C# Comparison
```csharp
// C# — classes are always reference types, so all methods can mutate
public void Advance(DeployStatus next)
{
    this.Status = next;  // always modifies the original
}
```

### Java Comparison
```java
// Java — same as C#, objects are always references
public void advance(DeployStatus next) {
    this.status = next;  // always modifies the original
}
```

### The "this" vs Named Receiver

| Language | Access instance | Explicit? |
|----------|----------------|-----------|
| Go | `d.Status` (named receiver `d`) | Yes — you choose the name |
| C# | `this.Status` | No — `this` is implicit |
| Java | `this.status` | No — `this` is implicit |

Go convention: use a short 1-2 letter abbreviation of the type name (`d` for Deployment, `env` for Environment, `s` for DeployStatus).

---

## Deep Dive: Factory Functions (Go's "Constructors")

Go has no `new` keyword for custom initialization. By convention, use `New*` functions:

```go
func NewDeployment(id, serviceName, imageTag, triggeredBy string, env Environment) *Deployment {
    now := time.Now()
    return &Deployment{
        ID:          id,
        ServiceName: serviceName,
        ImageTag:    imageTag,
        Environment: env,
        Status:      DeployStatusPending,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
}
```

The `&` operator returns a **pointer** to the struct (allocates on heap).

### C# Equivalent
```csharp
public Deployment(string id, string serviceName, string imageTag, string triggeredBy, Environment env)
{
    Id = id;
    ServiceName = serviceName;
    ImageTag = imageTag;
    Environment = env;
    Status = DeployStatus.Pending;
    CreatedAt = DateTime.UtcNow;
}
// Usage: var d = new Deployment("123", "api", "v1.0", "webhook", env);
```

### Java Equivalent
```java
public Deployment(String id, String serviceName, String imageTag, String triggeredBy, Environment env) {
    this.id = id;
    this.serviceName = serviceName;
    this.imageTag = imageTag;
    this.environment = env;
    this.status = DeployStatus.PENDING;
    this.createdAt = LocalDateTime.now();
}
// Usage: Deployment d = new Deployment("123", "api", "v1.0", "webhook", env);
```

---

## Deep Dive: Zero Values

Every type in Go has a **zero value** — no `null` surprises for value types.

| Type | Zero Value | C# Default | Java Default |
|------|-----------|------------|--------------|
| `string` | `""` (empty) | `""` | `null` ⚠️ |
| `int` | `0` | `0` | `0` |
| `bool` | `false` | `false` | `false` |
| `time.Time` | Zero time | `DateTime.MinValue` | `null` ⚠️ |
| `*Deployment` (pointer) | `nil` | `null` | `null` |
| `Deployment` (struct) | All fields zero | N/A (classes are ref) | N/A |

```go
var d Deployment  // All fields are zero-valued — no NullPointerException!
fmt.Println(d.Status)      // 0 (DeployStatusPending, since it's the first iota)
fmt.Println(d.ServiceName) // "" (empty string, not nil/null)
```

This is why Go has far fewer nil-pointer crashes than Java.

---

## Deep Dive: Composition (No Inheritance)

Go has **no inheritance**. Period. It uses composition — embedding structs inside other structs:

```go
type Deployment struct {
    Environment Environment  // Deployment HAS an Environment
    // NOT: Deployment extends Environment
}
```

Access embedded fields directly:
```go
d := NewDeployment("1", "api", "v1", "webhook", env)
fmt.Println(d.Environment.Name)         // "production"
fmt.Println(d.Environment.IsProduction) // true
```

### Why no inheritance?
Go's creators (from Google) found that deep inheritance hierarchies cause more problems than they solve. Composition is more flexible, more testable, and easier to understand.

> "Favor composition over inheritance" — Gang of Four (1994)
> Go just enforces it at the language level.

---

## Try It

```go
// In main.go or a test file:
env := domain.NewEnvironment("production", "eastus", "https://myapp.azurecontainerapps.io")
deploy := domain.NewDeployment("deploy-001", "payments-api", "v2.4.1", "github-webhook", env)

fmt.Println(deploy.Status)                // "pending"
fmt.Println(deploy.ShouldRequireApproval()) // true (production env)

deploy.Advance(domain.DeployStatusBuilding)
fmt.Println(deploy.Status)                // "building"
```

---

## Key Takeaways

1. **Structs = Go's classes** — but no inheritance, no constructors, no `this`
2. **`iota`** = Go's enum pattern — type-safe, auto-incrementing constants
3. **Value receiver** = reads only (copy). **Pointer receiver** = can mutate (reference)
4. **`New*` functions** = constructors by convention
5. **Zero values** = every type has a safe default (fewer nil panics than Java)
6. **Composition over inheritance** = enforced by the language

---

## Next: [Lesson 03 — Interfaces (Implicit Implementation)](./03-interfaces.md)
We'll define ports (interfaces) that the domain expects — without ever saying "implements".
