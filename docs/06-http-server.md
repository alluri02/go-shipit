# Lesson 06: HTTP Server (`net/http`)

## What We Built
```
go-shipit/
└── internal/
    └── transport/
        └── http/
            ├── server.go       ← Server struct, router setup, helpers
            └── handlers.go     ← Handler functions for each endpoint
```

### Endpoints:
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/deploys` | Start a new deployment |
| GET | `/deploys/{id}` | Get deployment by ID |
| GET | `/deploys?service=name` | List deployments for a service |

---

## The Core Difference

| | Go (`net/http`) | C# (ASP.NET Core) | Java (Spring Boot) |
|-|-----------------|--------------------|--------------------|
| **Framework** | Standard library (no framework) | ASP.NET Core (Microsoft) | Spring Boot (Pivotal) |
| **Router** | `http.ServeMux` (built-in) | Endpoint routing | DispatcherServlet |
| **Handler signature** | `func(w, r)` | `Func<HttpContext, Task>` | `@GetMapping` method |
| **Startup** | `http.ListenAndServe()` | `app.Run()` | `SpringApplication.run()` |
| **JSON** | `encoding/json` (stdlib) | `System.Text.Json` | Jackson |
| **Dependencies** | Zero external | Microsoft.AspNetCore.* | spring-boot-starter-web |

---

## Go HTTP Handler Pattern

Every HTTP handler in Go has this signature:

```go
func(w http.ResponseWriter, r *http.Request)
```

- **`w`** — you write the response to this (status code, headers, body)
- **`r`** — the incoming request (method, URL, headers, body)

### C# Equivalent (Minimal API):
```csharp
app.MapPost("/deploys", async (HttpContext ctx) => {
    var req = await ctx.Request.ReadFromJsonAsync<StartDeployRequest>();
    // ... process ...
    ctx.Response.StatusCode = 201;
    await ctx.Response.WriteAsJsonAsync(response);
});
```

### Java Equivalent (Spring):
```java
@PostMapping("/deploys")
public ResponseEntity<DeployResponse> startDeploy(@RequestBody StartDeployRequest req) {
    var deploy = service.startDeploy(req);
    return ResponseEntity.created(uri).body(toResponse(deploy));
}
```

### Go:
```go
func (s *Server) handleStartDeploy(w http.ResponseWriter, r *http.Request) {
    var req startDeployRequest
    if err := readJSON(r, &req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    // ... process ...
    writeJSON(w, http.StatusCreated, response)
}
```

---

## Deep Dive: Routing (Go 1.22+)

Go 1.22 added method+path patterns to the standard mux:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /health", s.handleHealth)
mux.HandleFunc("POST /deploys", s.handleStartDeploy)
mux.HandleFunc("GET /deploys/{id}", s.handleGetDeploy)
```

### Path parameters:
```go
id := r.PathValue("id")  // Go 1.22+
```

### Before Go 1.22 you needed external routers (gorilla/mux, chi). Now the stdlib is enough.

### C# equivalent:
```csharp
app.MapGet("/deploys/{id}", (string id) => ...);
```

### Java equivalent:
```java
@GetMapping("/deploys/{id}")
public DeployResponse getDeploy(@PathVariable String id) { ... }
```

---

## Deep Dive: JSON Serialization

### Struct Tags
```go
type startDeployRequest struct {
    ServiceName string `json:"service_name"`  // JSON field name mapping
    ImageTag    string `json:"image_tag"`
}
```

| Go struct tag | C# equivalent | Java equivalent |
|---------------|---------------|-----------------|
| `` `json:"service_name"` `` | `[JsonPropertyName("service_name")]` | `@JsonProperty("service_name")` |
| `` `json:"-"` `` | `[JsonIgnore]` | `@JsonIgnore` |
| `` `json:"name,omitempty"` `` | No direct equivalent | `@JsonInclude(NON_NULL)` |

### Encoding/Decoding:
```go
// Decode request body
json.NewDecoder(r.Body).Decode(&req)

// Encode response
json.NewEncoder(w).Encode(response)
```

### C# equivalent:
```csharp
var req = await JsonSerializer.DeserializeAsync<T>(stream);
await JsonSerializer.SerializeAsync(stream, response);
```

### Java equivalent:
```java
ObjectMapper mapper = new ObjectMapper();
T req = mapper.readValue(inputStream, T.class);
mapper.writeValue(outputStream, response);
```

---

## Deep Dive: Server Configuration

```go
s.server = &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,   // Max time to read request
    WriteTimeout: 10 * time.Second,  // Max time to write response
    IdleTimeout:  60 * time.Second,  // Max time for keep-alive connections
}
```

### Why timeouts matter:
Without them, a slow client can hold a connection forever → resource exhaustion → DoS.

### C# equivalent:
```csharp
builder.WebHost.ConfigureKestrel(options => {
    options.Limits.RequestHeadersTimeout = TimeSpan.FromSeconds(5);
    options.Limits.KeepAliveTimeout = TimeSpan.FromSeconds(60);
});
```

### Java equivalent:
```yaml
# application.yml
server:
  connection-timeout: 5000
  tomcat:
    keep-alive-timeout: 60000
```

---

## Deep Dive: Error → HTTP Status Mapping

The transport layer converts domain errors to HTTP status codes:

```go
func (s *Server) handleGetDeploy(w http.ResponseWriter, r *http.Request) {
    deploy, err := s.service.GetDeploy(id)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            writeError(w, http.StatusNotFound, "not found")  // 404
            return
        }
        writeError(w, http.StatusInternalServerError, "internal error")  // 500
        return
    }
    writeJSON(w, http.StatusOK, toDeployResponse(deploy))  // 200
}
```

| Domain Error | HTTP Status | C# | Java |
|-------------|-------------|-----|------|
| `ErrNotFound` | 404 | `throw new NotFoundException()` → mapped by middleware | `throw new ResponseStatusException(NOT_FOUND)` |
| `ValidationError` | 400 | `return Results.BadRequest()` | `return ResponseEntity.badRequest()` |
| Any other | 500 | `return Results.Problem()` | `return ResponseEntity.internalServerError()` |

---

## Try It

### Start the server:
```bash
go build -o shipit.exe ./cmd/shipit
.\shipit.exe serve
```

### In another terminal, test the endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Create a deployment
curl -X POST http://localhost:8080/deploys -H "Content-Type: application/json" -d '{
  "service_name": "payments-api",
  "image_tag": "v2.4.1",
  "triggered_by": "api",
  "environment": "staging",
  "region": "eastus"
}'

# Get it back (use the id from the response above)
curl http://localhost:8080/deploys/<id-from-above>

# List deploys for a service
curl "http://localhost:8080/deploys?service=payments-api"

# Test 404
curl http://localhost:8080/deploys/nonexistent
```

---

## Key Takeaways

1. **`net/http` is production-ready** — no external framework needed in Go.
2. **Handler signature**: `func(w http.ResponseWriter, r *http.Request)` — always.
3. **Go 1.22+ patterns**: `"GET /deploys/{id}"` — method + path in one string.
4. **JSON via struct tags** — `` `json:"field_name"` `` controls serialization.
5. **Always set timeouts** — prevents resource exhaustion.
6. **Transport maps errors to HTTP** — domain stays clean, HTTP layer does translation.

---

## Next: [Lesson 07 — Goroutines & Channels](./07-goroutines-channels.md)
We'll build the webhookprocessor with concurrent workers using goroutines and channels.
