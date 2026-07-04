# Lesson 13: AI Integration (GitHub Models)

## What We Built
```
go-shipit/
└── internal/
    └── adapters/
        └── ai/
            ├── risk_analyzer.go       ← HTTP client calling GPT-4o
            └── risk_analyzer_test.go  ← Tests using httptest mock server
```

---

## The Core Concept

Call an external AI API (OpenAI-compatible) from Go using:
- `net/http` client (same package as server!)
- `context.Context` for timeouts/cancellation
- `encoding/json` for request/response marshaling
- `httptest.NewServer` for testing without a real AI API

---

## Architecture: Risk Scoring in ShipIt

```
Deploy request arrives
       ↓
DeployService.StartDeploy()
       ↓
RiskAnalyzer.Assess(ctx, deploy, diff)
       ↓
HTTP POST → GitHub Models (GPT-4o)
       ↓
AI returns: {"score": 7, "reason": "production + DB migration"}
       ↓
if score >= 7 → require approval
if score < 7  → proceed automatically
```

---

## Pattern 1: HTTP Client with Context

```go
// Create request WITH context — enables timeout/cancellation
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Content-Type", "application/json")

// Send — automatically cancelled if ctx expires
resp, err := client.Do(req)
if err != nil {
    // Could be timeout, cancellation, or network error
    return err
}
defer resp.Body.Close()  // Always close!
```

### C# Equivalent:
```csharp
var request = new HttpRequestMessage(HttpMethod.Post, url) {
    Content = JsonContent.Create(body)
};
var response = await _httpClient.SendAsync(request, cancellationToken);
response.EnsureSuccessStatusCode();
var result = await response.Content.ReadFromJsonAsync<T>(cancellationToken);
```

### Java Equivalent:
```java
HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create(url))
    .header("Authorization", "Bearer " + token)
    .POST(HttpRequest.BodyPublishers.ofString(json))
    .build();
HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
```

---

## Pattern 2: Always Set HTTP Client Timeout

```go
// ✗ DANGEROUS — no timeout, can hang forever
client := http.DefaultClient

// ✓ SAFE — always create your own client with a timeout
client := &http.Client{
    Timeout: 30 * time.Second,
}
```

### Why?
Without a timeout, a slow/unresponsive external API can hold your goroutine forever, eventually exhausting all available goroutines → your service hangs.

### C# Equivalent:
```csharp
services.AddHttpClient<RiskAnalyzer>(client => {
    client.Timeout = TimeSpan.FromSeconds(30);
});
```

### Java Equivalent:
```java
HttpClient client = HttpClient.newBuilder()
    .connectTimeout(Duration.ofSeconds(5))
    .build();
```

---

## Pattern 3: Testing with httptest

```go
// Create a REAL HTTP server that returns controlled responses
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Verify request
    assert(r.Method == "POST")
    assert(r.Header.Get("Authorization") == "Bearer test-key")

    // Return mock response
    json.NewEncoder(w).Encode(mockResponse)
}))
defer server.Close()

// Point your client at the test server
analyzer := ai.NewRiskAnalyzer(server.URL, "test-key", "gpt-4o")
score, reason, err := analyzer.Assess(ctx, deploy, diff)
```

### Why this is great:
- Tests real HTTP behavior (headers, status codes, timeouts)
- No mocking framework
- Fast (in-process, no network)
- Tests error handling (return 500, close connection, etc.)

### C# Equivalent:
```csharp
var handler = new MockHttpMessageHandler();
handler.When(HttpMethod.Post, "*")
    .Respond("application/json", jsonResponse);
var client = handler.ToHttpClient();
```

### Java Equivalent:
```java
MockWebServer server = new MockWebServer();
server.enqueue(new MockResponse()
    .setBody(jsonResponse)
    .addHeader("Content-Type", "application/json"));
```

---

## Pattern 4: Structured JSON Communication

```go
// Request (Go struct → JSON)
type chatRequest struct {
    Model    string    `json:"model"`
    Messages []message `json:"messages"`
}

body, _ := json.Marshal(request)
// → {"model":"gpt-4o","messages":[...]}

// Response (JSON → Go struct)
var response chatResponse
json.NewDecoder(resp.Body).Decode(&response)
```

### The full cycle:
1. Define Go structs with `json:"field_name"` tags
2. `json.Marshal` → serialize to JSON bytes
3. Send via HTTP
4. Receive response
5. `json.Decode` → deserialize back to Go struct

---

## Pattern 5: Prompt Engineering for Structured Output

```go
systemPrompt := `You are a deployment risk assessor.
Respond with ONLY a JSON object: {"score": <1-10>, "reason": "<brief explanation>"}`
```

### Tips for reliable AI integration:
1. **Request JSON output** — easier to parse than free text
2. **Low temperature** (0.1-0.3) — more consistent/deterministic
3. **Validate output** — AI might return invalid JSON, always handle errors
4. **Clamp values** — ensure score is within expected range
5. **Set timeouts** — AI APIs can be slow (2-30 seconds)

---

## How It Fits in ShipIt

```go
// In the deploy flow (future: wired into DeployService)
analyzer := ai.NewRiskAnalyzer(
    "https://models.inference.ai.azure.com/chat/completions",
    os.Getenv("GITHUB_TOKEN"),
    "gpt-4o",
)

score, reason, err := analyzer.Assess(ctx, deploy, diffSummary)
if err != nil {
    log.Printf("AI risk assessment failed: %v (proceeding with default)", err)
    score = 5  // Default to medium risk if AI is unavailable
}

deploy.RiskScore = score
if deploy.IsHighRisk() {
    // Require human approval via Slack
    notifier.Notify("#deploys", fmt.Sprintf("⚠️ High risk (%d/10): %s\nReason: %s", score, deploy.ServiceName, reason))
}
```

---

## Try It

```bash
# Run the tests (uses httptest — no real AI API needed)
go test ./internal/adapters/ai -v

# To test with real GitHub Models:
$env:GITHUB_TOKEN = "your-github-token"
# (Integration code to be wired in a future lesson)
```

---

## Key Takeaways

1. **`http.NewRequestWithContext`** — always pass context to external calls.
2. **`&http.Client{Timeout: 30s}`** — NEVER use `DefaultClient` in production.
3. **`defer resp.Body.Close()`** — always close response bodies.
4. **`httptest.NewServer`** — test HTTP integrations without mocking frameworks.
5. **JSON round-trip** — `json.Marshal` → send → receive → `json.Decode`.
6. **Fail gracefully** — if AI is unavailable, use a default score, don't crash.
7. **Validate AI output** — it might return garbage. Always parse defensively.

---

## Next: [Lesson 14 — CLI with Cobra](./14-cli.md)
We'll replace our manual switch/case with a proper CLI framework for subcommand routing.
