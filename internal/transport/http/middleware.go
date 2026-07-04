package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

// WithTimeout is middleware that adds a timeout to each request's context.
//
// KEY CONCEPT: context.WithTimeout creates a new context that automatically
// cancels after the specified duration. If a handler takes too long,
// the context is cancelled → downstream calls (DB, HTTP, queue) abort.
//
// C# equivalent:
//   app.Use(async (ctx, next) => {
//       using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(10));
//       ctx.RequestAborted = cts.Token;
//       await next(ctx);
//   });
//
// Java equivalent:
//   @Bean
//   public WebFilter timeoutFilter() {
//       return (exchange, chain) -> chain.filter(exchange)
//           .timeout(Duration.ofSeconds(10));
//   }
func WithTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new context with timeout, derived from the request context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel() // Always cancel to release resources

			// Replace the request with one carrying the new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// WithRequestID is middleware that adds a unique request ID to the context.
//
// This enables distributed tracing — every log line includes the request ID.
//
// C# equivalent:
//   app.Use(async (ctx, next) => {
//       ctx.TraceIdentifier = Guid.NewGuid().ToString();
//       await next(ctx);
//   });
//
// Java equivalent:
//   MDC.put("requestId", UUID.randomUUID().toString());  // SLF4J MDC
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateID()
		}

		// Store in context — downstream handlers can retrieve it
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add to response headers for traceability
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

// WithLogging is middleware that logs each request with duration.
//
// C# equivalent:
//   app.UseHttpLogging();
//   // or custom: app.Use(async (ctx, next) => { var sw = Stopwatch.StartNew(); ... });
//
// Java equivalent:
//   @Component
//   public class LoggingFilter extends OncePerRequestFilter { ... }
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		reqID := GetRequestID(r.Context())

		log.Printf("[%s] %s %s %d %s",
			reqID, r.Method, r.URL.Path, wrapped.status, duration)
	})
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// --- Context Keys & Helpers ---

// contextKey is an unexported type for context keys to prevent collisions.
//
// Go best practice: context keys should be unexported types.
// This prevents other packages from accidentally overwriting your values.
type contextKey string

const requestIDKey contextKey = "request_id"

// GetRequestID extracts the request ID from context.
//
// C# equivalent: HttpContext.TraceIdentifier
// Java equivalent: MDC.get("requestId")
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// generateID creates a simple unique ID (timestamp-based for demo).
// In production, use github.com/google/uuid.
func generateID() string {
	return time.Now().Format("20060102-150405.000")
}
