package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alluri02/go-shipit/internal/domain"
)

// Server is the HTTP transport layer for ShipIt.
//
// Go's standard library `net/http` is production-ready — used at scale by Google, Cloudflare, etc.
// No framework needed (no Gin, Echo, Fiber). The stdlib is enough.
//
// C# equivalent: ASP.NET Core's WebApplication + Minimal APIs
//   var app = WebApplication.CreateBuilder(args).Build();
//   app.MapGet("/deploys/{id}", (string id) => ...);
//   app.Run();
//
// Java equivalent: Spring Boot's embedded Tomcat
//   @RestController
//   public class DeployController { ... }
//   SpringApplication.run(App.class, args);
type Server struct {
	service *domain.DeployService
	server  *http.Server
}

// NewServer creates an HTTP server with all routes registered.
// Dependencies are injected via constructor (Lesson 05 pattern).
func NewServer(addr string, service *domain.DeployService) *Server {
	s := &Server{service: service}

	// Create a mux (router) — maps URL patterns to handler functions.
	//
	// C# equivalent: app.MapGet("/path", handler)
	// Java equivalent: @GetMapping("/path")
	mux := http.NewServeMux()

	// Register routes — Go 1.22+ supports method + path patterns
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /deploys", s.handleStartDeploy)
	mux.HandleFunc("GET /deploys/{id}", s.handleGetDeploy)
	mux.HandleFunc("GET /deploys", s.handleListDeploys)

	// Apply middleware chain (Lesson 11)
	// Middleware wraps the handler — executed in reverse order (outermost first).
	//
	// C# equivalent: app.UseHttpLogging(); app.UseRouting(); ...
	// Java equivalent: @Order(1) Filter, @Order(2) Filter, ...
	var handler http.Handler = mux
	handler = WithTimeout(10 * time.Second)(handler)
	handler = WithLogging(handler)
	handler = WithRequestID(handler)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start begins listening. Blocks until the server stops.
//
// C# equivalent: await app.RunAsync();
// Java equivalent: SpringApplication.run(App.class, args); (blocks)
func (s *Server) Start() error {
	log.Printf("HTTP server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// --- Helper functions for HTTP responses ---

// writeJSON writes a JSON response with the given status code.
//
// C# equivalent: return Results.Ok(obj);  or  return Results.Json(obj, statusCode: 201);
// Java equivalent: return ResponseEntity.status(201).body(obj);
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

// writeError writes an error response as JSON.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// readJSON decodes a JSON request body into the target struct.
//
// C# equivalent: [FromBody] in controller parameters (auto-deserialization)
// Java equivalent: @RequestBody annotation (auto-deserialization)
func readJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
