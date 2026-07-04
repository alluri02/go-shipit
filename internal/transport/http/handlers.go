package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alluri02/go-shipit/internal/domain"
)

// --- Health Check ---

// handleHealth returns service status. Every production service needs this.
//
// C# equivalent:
//   app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
//
// Java equivalent:
//   @GetMapping("/health")
//   public Map<String, String> health() { return Map.of("status", "ok"); }
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "shipit",
		"version": domain.Version,
	})
}

// --- Start Deploy ---

// startDeployRequest is the expected JSON body for POST /deploys.
//
// The `json:"..."` tags control JSON serialization — field name mapping.
//
// C# equivalent:
//   public record StartDeployRequest(string ServiceName, string ImageTag, ...);
//
// Java equivalent:
//   public record StartDeployRequest(String serviceName, String imageTag, ...) {}
type startDeployRequest struct {
	ServiceName string `json:"service_name"`
	ImageTag    string `json:"image_tag"`
	TriggeredBy string `json:"triggered_by"`
	Environment string `json:"environment"`
	Region      string `json:"region"`
}

// deployResponse is the JSON response for a deployment.
type deployResponse struct {
	ID          string `json:"id"`
	ServiceName string `json:"service_name"`
	ImageTag    string `json:"image_tag"`
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Region      string `json:"region"`
	TriggeredBy string `json:"triggered_by"`
	RiskScore   int    `json:"risk_score"`
	CreatedAt   string `json:"created_at"`
}

// handleStartDeploy creates a new deployment.
//
// In Go, HTTP handlers have a standard signature:
//   func(w http.ResponseWriter, r *http.Request)
//
// C# equivalent:
//   app.MapPost("/deploys", async ([FromBody] StartDeployRequest req, DeployService svc) => {
//       var deploy = await svc.StartDeployAsync(req);
//       return Results.Created($"/deploys/{deploy.Id}", deploy);
//   });
//
// Java equivalent:
//   @PostMapping("/deploys")
//   public ResponseEntity<DeployResponse> startDeploy(@RequestBody StartDeployRequest req) {
//       var deploy = deployService.startDeploy(req);
//       return ResponseEntity.created(URI.create("/deploys/" + deploy.getId())).body(deploy);
//   }
func (s *Server) handleStartDeploy(w http.ResponseWriter, r *http.Request) {
	var req startDeployRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Generate a simple ID (in production, use UUID)
	id := fmt.Sprintf("deploy-%d", time.Now().UnixNano())

	env := domain.NewEnvironment(req.Environment, req.Region, "")

	deploy, err := s.service.StartDeploy(id, req.ServiceName, req.ImageTag, req.TriggeredBy, env)
	if err != nil {
		// Map domain errors to HTTP status codes
		var valErr *domain.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, toDeployResponse(deploy))
}

// --- Get Deploy ---

// handleGetDeploy retrieves a single deployment by ID.
//
// Go 1.22+ path parameters use {name} syntax in the pattern,
// accessed via r.PathValue("name").
//
// C# equivalent: app.MapGet("/deploys/{id}", (string id) => ...);
// Java equivalent: @GetMapping("/deploys/{id}") ... @PathVariable String id
func (s *Server) handleGetDeploy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Go 1.22+ path parameter extraction

	deploy, err := s.service.GetDeploy(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, fmt.Sprintf("deployment %q not found", id))
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, toDeployResponse(deploy))
}

// --- List Deploys ---

// handleListDeploys returns recent deployments for a service.
//
// Query parameters in Go: r.URL.Query().Get("key")
//
// C# equivalent: app.MapGet("/deploys", ([FromQuery] string service) => ...);
// Java equivalent: @GetMapping("/deploys") ... @RequestParam String service
func (s *Server) handleListDeploys(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'service' is required")
		return
	}

	deploys, err := s.service.ListDeploys(serviceName, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var responses []deployResponse
	for _, d := range deploys {
		responses = append(responses, toDeployResponse(d))
	}

	writeJSON(w, http.StatusOK, responses)
}

// --- Helpers ---

func toDeployResponse(d *domain.Deployment) deployResponse {
	return deployResponse{
		ID:          d.ID,
		ServiceName: d.ServiceName,
		ImageTag:    d.ImageTag,
		Status:      d.Status.String(),
		Environment: d.Environment.Name,
		Region:      d.Environment.Region,
		TriggeredBy: d.TriggeredBy,
		RiskScore:   d.RiskScore,
		CreatedAt:   d.CreatedAt.Format(time.RFC3339),
	}
}
