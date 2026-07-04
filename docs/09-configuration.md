# Lesson 09: Configuration & Environment

## What We Built
```
go-shipit/
└── internal/
    └── config/
        ├── config.go           ← Config struct + Load() + Validate()
        └── config_test.go      ← Tests for config loading
```

---

## The Core Difference

| | Go | C# | Java |
|-|-----|-----|------|
| **Config source** | `os.Getenv()` (env vars) | `appsettings.json` + env vars + secrets | `application.yml` + env vars |
| **Framework** | None (stdlib) | `IConfiguration` + `IOptions<T>` | `@ConfigurationProperties` |
| **DI integration** | Manual (pass config struct) | `services.Configure<T>()` | Spring auto-binds |
| **Validation** | Manual (write a `Validate()` method) | Data annotations / FluentValidation | `@Validated` + JSR-303 |
| **Secrets** | Env vars / secret manager | Azure Key Vault / User Secrets | Spring Vault / env vars |

---

## The 12-Factor App Way

[12factor.net/config](https://12factor.net/config): **Store config in environment variables.**

```bash
# Set config via env vars (works everywhere: local, Docker, K8s, Azure)
export SHIPIT_SERVER_ADDR=":9090"
export SHIPIT_ENV="production"
export SHIPIT_DATABASE_URL="mysql://user:pass@host/db"
export SHIPIT_WORKER_COUNT="5"
```

---

## Go Pattern: Config Struct + Load Function

```go
type Config struct {
    ServerAddr  string
    DatabaseURL string
    WorkerCount int
    Env         string
}

func Load() *Config {
    return &Config{
        ServerAddr:  getEnv("SHIPIT_SERVER_ADDR", ":8080"),
        DatabaseURL: getEnv("SHIPIT_DATABASE_URL", ""),
        WorkerCount: getEnvInt("SHIPIT_WORKER_COUNT", 3),
        Env:         getEnv("SHIPIT_ENV", "development"),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

### C# Equivalent:
```csharp
// appsettings.json
{
    "ShipIt": {
        "ServerAddr": ":8080",
        "DatabaseURL": "",
        "WorkerCount": 3,
        "Env": "development"
    }
}

// Startup.cs
services.Configure<ShipItConfig>(config.GetSection("ShipIt"));

// Usage (injected)
public class MyService {
    public MyService(IOptions<ShipItConfig> options) {
        var cfg = options.Value;
    }
}
```

### Java Equivalent:
```yaml
# application.yml
shipit:
  server-addr: ${SHIPIT_SERVER_ADDR::8080}
  database-url: ${SHIPIT_DATABASE_URL:}
  worker-count: ${SHIPIT_WORKER_COUNT:3}
  env: ${SHIPIT_ENV:development}
```

```java
@ConfigurationProperties(prefix = "shipit")
public class ShipItConfig {
    private String serverAddr;
    private String databaseUrl;
    private int workerCount;
    private String env;
    // getters/setters
}
```

---

## Deep Dive: Environment-Specific Validation

```go
func (c *Config) Validate() error {
    var missing []string

    if c.IsProduction() {
        if c.DatabaseURL == "" {
            missing = append(missing, "SHIPIT_DATABASE_URL")
        }
        if c.SlackToken == "" {
            missing = append(missing, "SHIPIT_SLACK_TOKEN")
        }
    }

    if len(missing) > 0 {
        return fmt.Errorf("missing required config: %v", missing)
    }
    return nil
}
```

- **Development**: runs with defaults (no DB, no queue needed)
- **Production**: fails fast if critical config is missing

### C# Equivalent:
```csharp
// Using FluentValidation
public class ConfigValidator : AbstractValidator<ShipItConfig> {
    public ConfigValidator(IHostEnvironment env) {
        if (env.IsProduction()) {
            RuleFor(c => c.DatabaseURL).NotEmpty();
            RuleFor(c => c.SlackToken).NotEmpty();
        }
    }
}
```

---

## Deep Dive: Never Log Secrets

```go
func (c *Config) String() string {
    return fmt.Sprintf("Config{env=%s, addr=%s, db=%s}",
        c.Env, c.ServerAddr, maskSecret(c.DatabaseURL))
}

func maskSecret(s string) string {
    if s == "" { return "<not set>" }
    if len(s) <= 8 { return "****" }
    return s[:4] + "****" + s[len(s)-4:]
}
```

Output: `Config{env=production, addr=:8080, db=mysq****t/db}`

### NEVER do this:
```go
log.Printf("Starting with config: %+v", cfg)  // LEAKS ALL SECRETS!
```

### ALWAYS do this:
```go
log.Printf("Starting with config: %s", cfg)   // Uses safe String() method
```

---

## Deep Dive: Testing Config

```go
func TestLoad_FromEnv(t *testing.T) {
    os.Setenv("SHIPIT_SERVER_ADDR", ":9090")
    defer os.Unsetenv("SHIPIT_SERVER_ADDR")  // Clean up after test

    cfg := config.Load()

    if cfg.ServerAddr != ":9090" {
        t.Errorf("got %q, want %q", cfg.ServerAddr, ":9090")
    }
}
```

### The `defer` pattern for cleanup:
```go
os.Setenv("KEY", "value")
defer os.Unsetenv("KEY")  // Guaranteed to run when function exits
```

C# equivalent: `try { ... } finally { Environment.SetEnvironmentVariable("KEY", null); }`
Java equivalent: `@AfterEach void cleanup() { ... }`

---

## Try It

```bash
# Default config (development)
go build -o shipit.exe ./cmd/shipit
.\shipit.exe serve

# Custom port via env var
$env:SHIPIT_SERVER_ADDR = ":9090"
.\shipit.exe serve

# Reset
Remove-Item Env:SHIPIT_SERVER_ADDR
```

---

## Key Takeaways

1. **Env vars are the standard** — 12-Factor App pattern. Works in Docker, K8s, Azure, everywhere.
2. **Config struct + Load()** — simple, testable, no framework.
3. **Defaults for dev, require for prod** — `Validate()` adapts to environment.
4. **Never log secrets** — implement `String()` with masking.
5. **`os.Getenv()`** — that's it. No `IConfiguration`, no `@Value`, no YAML parser.
6. **`defer`** — for cleanup in tests (like `finally` in C#/Java).

---

## Next: [Lesson 10 — Database (SQL, No ORM)](./10-database.md)
We'll connect to MySQL, write raw SQL queries, and manage schema with Skeema.
