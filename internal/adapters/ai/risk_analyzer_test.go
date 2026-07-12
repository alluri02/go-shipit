package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alluri02/go-shipit/internal/adapters/ai"
	"github.com/alluri02/go-shipit/internal/domain"
)

// TestRiskAnalyzer_Assess uses httptest to mock the AI API.
//
// KEY GO CONCEPT: httptest.NewServer — creates a real HTTP server for testing.
// No mocking framework needed — you run a real server that returns controlled responses.
//
// C# equivalent:
//   var mockHttp = new MockHttpMessageHandler();
//   mockHttp.When(HttpMethod.Post, "*").Respond("application/json", jsonResponse);
//   var client = mockHttp.ToHttpClient();
//
// Java equivalent:
//   MockWebServer server = new MockWebServer();
//   server.enqueue(new MockResponse().setBody(jsonResponse));
func TestRiskAnalyzer_Assess(t *testing.T) {
	// Create a fake AI API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected auth header, got %q", r.Header.Get("Authorization"))
		}

		// Return a mock AI response
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": `{"score": 7, "reason": "Production deployment with database migration"}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Create analyzer pointing to our mock server
	analyzer := ai.NewRiskAnalyzer(server.URL, "test-key", "gpt-4o")

	// Test
	env := domain.NewEnvironment("production", "eastus", "")
	deploy := domain.NewDeployment("d-001", "payments-api", "v2.4.1", "github-webhook", env)

	score, reason, err := analyzer.Assess(context.Background(), deploy, "Added retry logic for Stripe webhooks\nMigration: ALTER TABLE payments ADD COLUMN retry_count INT")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 7 {
		t.Errorf("score = %d, want 7", score)
	}
	if reason != "Production deployment with database migration" {
		t.Errorf("reason = %q, want expected message", reason)
	}
}

func TestRiskAnalyzer_Assess_APIError(t *testing.T) {
	// Server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	analyzer := ai.NewRiskAnalyzer(server.URL, "test-key", "gpt-4o")
	env := domain.NewEnvironment("staging", "eastus", "")
	deploy := domain.NewDeployment("d-001", "svc", "v1", "api", env)

	_, _, err := analyzer.Assess(context.Background(), deploy, "small change")

	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestRiskAnalyzer_Assess_ContextCancelled(t *testing.T) {
	// Server that never responds (simulates slow API)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // Block until cancelled
	}))
	defer server.Close()

	analyzer := ai.NewRiskAnalyzer(server.URL, "test-key", "gpt-4o")
	env := domain.NewEnvironment("staging", "eastus", "")
	deploy := domain.NewDeployment("d-001", "svc", "v1", "api", env)

	// Cancel the context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := analyzer.Assess(ctx, deploy, "")

	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
