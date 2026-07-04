# Lesson 08: Testing (Table-Driven)

## What We Built
```
go-shipit/
└── internal/
    └── domain/
        ├── deployment_test.go    ← Model tests (table-driven)
        ├── validate_test.go      ← Validation tests
        └── service_test.go       ← Service tests with mocks
```

---

## The Core Difference

| | Go | C# (xUnit) | Java (JUnit 5) |
|-|-----|-------------|----------------|
| **Framework** | Built-in (`testing` package) | xUnit / NUnit / MSTest | JUnit 5 / TestNG |
| **Test runner** | `go test` | `dotnet test` | `mvn test` / `gradle test` |
| **Assertions** | None built-in (`if` + `t.Errorf`) | `Assert.Equal(expected, actual)` | `assertEquals(expected, actual)` |
| **Parametrized** | Table-driven (slice of structs) | `[Theory] + [InlineData]` | `@ParameterizedTest + @CsvSource` |
| **Mocking** | Manual (just a struct) | Moq / NSubstitute | Mockito / EasyMock |
| **File naming** | `*_test.go` (same package) | Separate test project | Separate `src/test/` directory |
| **Test naming** | `func TestXxx(t *testing.T)` | `[Fact] public void Xxx()` | `@Test void xxx()` |

---

## Pattern 1: Simple Test

```go
func TestNewDeployment(t *testing.T) {
    deploy := domain.NewDeployment("d-001", "svc", "v1", "api", env)

    if deploy.ID != "d-001" {
        t.Errorf("ID = %q, want %q", deploy.ID, "d-001")
    }
}
```

### C# Equivalent:
```csharp
[Fact]
public void NewDeployment_SetsID() {
    var deploy = new Deployment("d-001", "svc", "v1", "api", env);
    Assert.Equal("d-001", deploy.Id);
}
```

### Java Equivalent:
```java
@Test
void newDeployment_setsId() {
    var deploy = new Deployment("d-001", "svc", "v1", "api", env);
    assertEquals("d-001", deploy.getId());
}
```

### Why Go has no assertions:
Go's philosophy: **`if` is clear enough.** When it fails, `t.Errorf` gives you a custom message. No need to learn assertion APIs.

---

## Pattern 2: Table-Driven Tests (Go's Signature Pattern)

```go
func TestIsHighRisk(t *testing.T) {
    tests := []struct {
        name  string
        score int
        want  bool
    }{
        {"low risk", 3, false},
        {"borderline", 6, false},
        {"high risk", 7, true},
        {"critical", 10, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            deploy := newDeploy()
            deploy.RiskScore = tt.score
            if got := deploy.IsHighRisk(); got != tt.want {
                t.Errorf("IsHighRisk() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### C# Equivalent (xUnit Theory):
```csharp
[Theory]
[InlineData(3, false)]
[InlineData(6, false)]
[InlineData(7, true)]
[InlineData(10, true)]
public void IsHighRisk_ReturnsExpected(int score, bool expected) {
    var deploy = new Deployment { RiskScore = score };
    Assert.Equal(expected, deploy.IsHighRisk());
}
```

### Java Equivalent (Parameterized):
```java
@ParameterizedTest
@CsvSource({"3,false", "6,false", "7,true", "10,true"})
void isHighRisk_returnsExpected(int score, boolean expected) {
    var deploy = new Deployment();
    deploy.setRiskScore(score);
    assertEquals(expected, deploy.isHighRisk());
}
```

### Why table-driven is better:
- Add new cases = add one line to the table
- Every case gets a name (shown in output)
- `t.Run` creates subtests you can run individually

---

## Pattern 3: Mocking (No Framework)

In Go, mocking is just... a struct with the right methods:

```go
type mockRepo struct {
    deployments map[string]*domain.Deployment
    saveErr     error  // inject failures
}

func (m *mockRepo) Save(d *domain.Deployment) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.deployments[d.ID] = d
    return nil
}
```

### C# Equivalent (Moq):
```csharp
var mockRepo = new Mock<IDeployRepository>();
mockRepo.Setup(r => r.SaveAsync(It.IsAny<Deployment>()))
        .ThrowsAsync(new DbException("connection lost"));
```

### Java Equivalent (Mockito):
```java
DeployRepository mockRepo = mock(DeployRepository.class);
when(mockRepo.save(any())).thenThrow(new DatabaseException("connection lost"));
```

### Why Go's approach works:
- No framework to learn
- No code generation
- Compile-time verified (if interface changes, mock breaks)
- Easy to add custom behavior (counters, captures, conditional errors)

---

## Pattern 4: Testing Errors

```go
func TestGetDeploy_NotFound(t *testing.T) {
    service := setupService()

    _, err := service.GetDeploy("nonexistent")

    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if !errors.Is(err, domain.ErrNotFound) {
        t.Errorf("expected ErrNotFound, got: %v", err)
    }
}
```

### Test error types:
```go
var valErr *domain.ValidationError
if !errors.As(err, &valErr) {
    t.Fatalf("expected ValidationError, got %T", err)
}
if valErr.Field != "serviceName" {
    t.Errorf("field = %q, want %q", valErr.Field, "serviceName")
}
```

---

## Key Testing Commands

```bash
# Run all tests
go test ./...

# Run with verbose output (see each test name)
go test ./... -v

# Run specific test by name
go test ./internal/domain -run TestDeployService

# Run specific subtest
go test ./internal/domain -run TestDeployment_AdvanceSafe/pending_to_building

# Run with coverage
go test ./... -cover

# Generate coverage HTML report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### C# Equivalents:
```bash
dotnet test
dotnet test --filter "FullyQualifiedName~DeployServiceTests"
dotnet test /p:CollectCoverage=true
```

### Java Equivalents:
```bash
mvn test
mvn test -Dtest=DeployServiceTest
mvn test jacoco:report
```

---

## Test File Organization

| Convention | Go | C# | Java |
|-----------|-----|-----|------|
| Location | Same directory as code | Separate `.Tests` project | `src/test/java/` mirror |
| File name | `thing_test.go` | `ThingTests.cs` | `ThingTest.java` |
| Package | Same or `_test` suffix | Separate namespace | Same package |
| Build | Excluded from binary automatically | Separate assembly | Separate compilation |

Go test files are **never compiled into the final binary** — they're excluded by the `_test.go` suffix convention.

---

## t.Fatal vs t.Error

| Method | Effect | C# Equivalent | Java Equivalent |
|--------|--------|---------------|-----------------|
| `t.Error(msg)` | Log + continue | `Assert.True(false, msg)` (non-fatal) | Multiple soft assertions |
| `t.Fatal(msg)` | Log + stop this test | `throw` in test | `fail()` |
| `t.Errorf(fmt, ...)` | Formatted error + continue | — | — |
| `t.Fatalf(fmt, ...)` | Formatted error + stop | — | — |

Use `t.Fatal` when continuing makes no sense (nil pointer would crash).
Use `t.Error` when you want to see ALL failures at once.

---

## Try It

```bash
go test ./... -v
```

Output shows each test and subtest:
```
=== RUN   TestDeployment_AdvanceSafe
=== RUN   TestDeployment_AdvanceSafe/pending_to_building
=== RUN   TestDeployment_AdvanceSafe/building_to_pushing
--- PASS: TestDeployment_AdvanceSafe (0.00s)
    --- PASS: TestDeployment_AdvanceSafe/pending_to_building (0.00s)
    --- PASS: TestDeployment_AdvanceSafe/building_to_pushing (0.00s)
```

---

## Key Takeaways

1. **`go test ./...`** — runs all tests in all packages. Built-in. No install needed.
2. **Table-driven tests** — slice of structs + `t.Run`. Go's #1 testing pattern.
3. **No assertion library** — `if` + `t.Errorf` is the Go way.
4. **Mocking is free** — just a struct with methods. No framework.
5. **`*_test.go`** — automatically excluded from production binary.
6. **`t.Run("name", ...)`** — creates subtests you can filter with `-run`.
7. **`-cover`** — built-in coverage with no extra tools.

---

## Next: [Lesson 09 — Configuration & Environment](./09-configuration.md)
We'll load config from environment variables and files — the 12-factor app way.
