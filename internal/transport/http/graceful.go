package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alluri02/go-shipit/internal/domain"
)

// GracefulServer wraps Server with graceful shutdown support.
//
// KEY GO CONCEPT: context.Context — a request-scoped value that carries:
//   1. Deadlines/timeouts
//   2. Cancellation signals
//   3. Request-scoped values (user ID, trace ID, etc.)
//
// Every function that does I/O or takes time should accept a context.Context
// as its FIRST parameter. This is Go's #1 convention.
//
// C# equivalent: CancellationToken
//   public async Task<Deploy> GetAsync(string id, CancellationToken ct) { ... }
//
// Java equivalent: No direct equivalent. Closest:
//   - CompletableFuture cancellation
//   - Thread.interrupt()
//   - Reactor's Mono.timeout()
type GracefulServer struct {
	server  *Server
	httpSrv *http.Server
}

// NewGracefulServer creates a server that handles OS signals for shutdown.
func NewGracefulServer(addr string, service *domain.DeployService) *GracefulServer {
	srv := NewServer(addr, service)
	return &GracefulServer{
		server:  srv,
		httpSrv: srv.server,
	}
}

// StartWithGracefulShutdown starts the HTTP server and handles SIGINT/SIGTERM.
//
// Graceful shutdown means:
//   1. Stop accepting NEW connections
//   2. Wait for IN-FLIGHT requests to complete (up to a timeout)
//   3. Then exit cleanly
//
// Without graceful shutdown, active requests get terminated mid-response.
//
// C# equivalent:
//   var app = builder.Build();
//   app.Lifetime.ApplicationStopping.Register(() => { /* cleanup */ });
//   await app.RunAsync();  // ASP.NET handles SIGTERM automatically
//
// Java equivalent:
//   Runtime.getRuntime().addShutdownHook(new Thread(() -> {
//       server.shutdown();
//       server.awaitTermination(30, TimeUnit.SECONDS);
//   }));
func (gs *GracefulServer) StartWithGracefulShutdown() error {
	// Channel to receive OS signals
	// os.Signal is sent when the process receives SIGINT (Ctrl+C) or SIGTERM (Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine (non-blocking)
	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", gs.httpSrv.Addr)
		if err := gs.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Block until we receive a signal OR a server error
	select {
	case sig := <-quit:
		log.Printf("Received signal: %v. Shutting down gracefully...", sig)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	// Create a context with timeout for shutdown
	// Give in-flight requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown gracefully — stops new connections, waits for active ones
	if err := gs.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
