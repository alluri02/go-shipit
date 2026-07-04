package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alluri02/go-shipit/internal/domain"
)

// RiskAnalyzer calls GitHub Models (GPT-4o) to assess deployment risk.
//
// KEY GO CONCEPT: HTTP client for external API calls.
// Go's net/http package works for both servers AND clients.
// No separate "HttpClient" class — the same package handles everything.
//
// C# equivalent:
//   public class RiskAnalyzer : IRiskAnalyzer {
//       private readonly HttpClient _http;
//       public RiskAnalyzer(HttpClient http) { _http = http; }
//       public async Task<RiskAssessment> AssessAsync(Deployment d, string diff) { ... }
//   }
//
// Java equivalent:
//   @Service
//   public class RiskAnalyzer {
//       private final WebClient webClient;
//       public RiskAnalyzer(WebClient.Builder builder) { this.webClient = builder.build(); }
//       public Mono<RiskAssessment> assess(Deployment d, String diff) { ... }
//   }
type RiskAnalyzer struct {
	client   *http.Client
	endpoint string // GitHub Models endpoint
	apiKey   string // GitHub token
	model    string // e.g., "gpt-4o"
}

// NewRiskAnalyzer creates a risk analyzer that calls GitHub Models.
//
// Note: We create our own http.Client with a timeout.
// NEVER use http.DefaultClient in production — it has no timeout!
//
// C# equivalent:
//   services.AddHttpClient<RiskAnalyzer>(client => {
//       client.BaseAddress = new Uri(endpoint);
//       client.Timeout = TimeSpan.FromSeconds(30);
//   });
//
// Java equivalent:
//   WebClient.builder()
//       .baseUrl(endpoint)
//       .defaultHeader("Authorization", "Bearer " + token)
//       .build();
func NewRiskAnalyzer(endpoint, apiKey, model string) *RiskAnalyzer {
	return &RiskAnalyzer{
		client: &http.Client{
			Timeout: 30 * time.Second, // ALWAYS set a timeout on HTTP clients
		},
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
	}
}

// Assess sends deployment context to the AI model and returns a risk score.
//
// Implements domain.RiskAnalyzer interface (implicitly — Lesson 03).
//
// KEY PATTERN: context.Context propagation
// The ctx parameter carries the request timeout. If the caller's context
// is cancelled (e.g., HTTP request timeout), this API call is also cancelled.
func (ra *RiskAnalyzer) Assess(ctx context.Context, deployment *domain.Deployment, diffSummary string) (score int, reason string, err error) {
	// Build the prompt
	prompt := buildRiskPrompt(deployment, diffSummary)

	// Create the request body (OpenAI-compatible API format)
	reqBody := chatRequest{
		Model: ra.model,
		Messages: []message{
			{
				Role:    "system",
				Content: "You are a deployment risk assessor. Analyze the deployment context and respond with ONLY a JSON object: {\"score\": <1-10>, \"reason\": \"<brief explanation>\"}. Score 1 = no risk, 10 = extreme risk.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2, // Low temperature = more deterministic/consistent
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return 0, "", fmt.Errorf("ai.Assess: marshal request: %w", err)
	}

	// Create HTTP request with context (Lesson 11)
	// If ctx is cancelled, the request is aborted automatically.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ra.endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, "", fmt.Errorf("ai.Assess: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ra.apiKey)

	// Send the request
	resp, err := ra.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("ai.Assess: request failed: %w", err)
	}
	defer resp.Body.Close() // Always close response bodies (like rows.Close in DB)

	// Check status code
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("ai.Assess: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return 0, "", fmt.Errorf("ai.Assess: decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return 0, "", fmt.Errorf("ai.Assess: no choices in response")
	}

	// Parse the AI's JSON response
	var result riskResult
	content := chatResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return 0, "", fmt.Errorf("ai.Assess: parse AI response %q: %w", content, err)
	}

	// Clamp score to valid range
	if result.Score < 1 {
		result.Score = 1
	}
	if result.Score > 10 {
		result.Score = 10
	}

	return result.Score, result.Reason, nil
}

// buildRiskPrompt assembles context for the AI to evaluate.
func buildRiskPrompt(d *domain.Deployment, diffSummary string) string {
	return fmt.Sprintf(`Assess the risk of this deployment:

Service: %s
Image Tag: %s
Environment: %s (region: %s)
Triggered By: %s
Is Production: %v

Diff Summary:
%s

Consider:
- Is this a production deployment?
- How large/complex are the changes?
- Is it a high-traffic service?
- Time of deployment (business hours vs off-hours)
- Any database migrations mentioned?`,
		d.ServiceName,
		d.ImageTag,
		d.Environment.Name,
		d.Environment.Region,
		d.TriggeredBy,
		d.Environment.IsProduction,
		diffSummary,
	)
}

// --- Request/Response types for OpenAI-compatible API ---

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}

type riskResult struct {
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}
