package http

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// WithAuth is middleware that validates API key authentication.
//
// KEY GO CONCEPT: Function composition via closures.
// Middleware returns a function that "closes over" the config (apiKey).
// This is Go's alternative to decorator pattern / DI-injected filters.
//
// C# equivalent:
//   app.UseMiddleware<ApiKeyAuthMiddleware>();
//   // or:
//   app.Use(async (ctx, next) => {
//       if (ctx.Request.Headers["X-API-Key"] != expectedKey)
//           { ctx.Response.StatusCode = 401; return; }
//       await next(ctx);
//   });
//
// Java equivalent:
//   @Component
//   public class ApiKeyFilter extends OncePerRequestFilter {
//       @Value("${api.key}") private String apiKey;
//       protected void doFilterInternal(...) {
//           if (!apiKey.equals(request.getHeader("X-API-Key")))
//               { response.sendError(401); return; }
//           filterChain.doFilter(request, response);
//       }
//   }
func WithAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check (always public)
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Check API key header
			key := r.Header.Get("X-API-Key")
			if key == "" {
				writeError(w, http.StatusUnauthorized, "missing X-API-Key header")
				return
			}
			if key != apiKey {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithRateLimit is middleware that limits requests per IP using a token bucket.
//
// KEY GO CONCEPT: sync.Mutex for protecting shared state across goroutines.
// Each incoming request is a goroutine — shared maps need locking.
//
// C# equivalent:
//   builder.Services.AddRateLimiter(options => {
//       options.AddFixedWindowLimiter("api", o => {
//           o.PermitLimit = 100;
//           o.Window = TimeSpan.FromMinutes(1);
//       });
//   });
//   app.UseRateLimiter();
//
// Java equivalent:
//   @Bean
//   public RateLimiter rateLimiter() {
//       return RateLimiter.create(100.0 / 60);  // Guava RateLimiter
//   }
func WithRateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	type client struct {
		tokens    int
		lastReset time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				c = &client{tokens: requestsPerMinute, lastReset: time.Now()}
				clients[ip] = c
			}

			// Reset tokens if a minute has passed
			if time.Since(c.lastReset) > time.Minute {
				c.tokens = requestsPerMinute
				c.lastReset = time.Now()
			}

			if c.tokens <= 0 {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			c.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// WithCORS is middleware that adds Cross-Origin Resource Sharing headers.
//
// C# equivalent: app.UseCors(policy => policy.AllowAnyOrigin().AllowAnyMethod());
// Java equivalent: @CrossOrigin or WebMvcConfigurer.addCorsMappings()
func WithCORS(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID")

			// Handle preflight OPTIONS requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithRecover is middleware that catches panics and returns 500 instead of crashing.
//
// KEY GO CONCEPT: recover() — catches panics (like catch in C#/Java).
// Panics are Go's "exceptions" — but only for truly exceptional cases (bugs, nil dereference).
// Normal errors use return values (Lesson 04). Panics are for programmer errors.
//
// C# equivalent:
//   app.UseExceptionHandler(error => error.Run(async ctx => {
//       ctx.Response.StatusCode = 500;
//       await ctx.Response.WriteAsync("Internal Server Error");
//   }));
//
// Java equivalent:
//   @ExceptionHandler(Exception.class)
//   public ResponseEntity<String> handleException(Exception ex) {
//       return ResponseEntity.status(500).body("Internal Server Error");
//   }
func WithRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v (request: %s %s)", err, r.Method, r.URL.Path)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// WithStructuredLogging replaces the basic logging middleware with JSON-formatted logs.
//
// Structured logging = logs as JSON objects (machine-parseable).
// Essential for production — enables searching in tools like Grafana/Loki, Datadog, Azure Monitor.
//
// C# equivalent: Serilog with JSON formatter
//   Log.Logger = new LoggerConfiguration()
//       .WriteTo.Console(new JsonFormatter())
//       .CreateLogger();
//
// Java equivalent: Logback with JSON encoder
//   <encoder class="net.logstash.logback.encoder.LogstashEncoder"/>
func WithStructuredLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		reqID := GetRequestID(r.Context())

		// Structured log entry as JSON
		entry := map[string]any{
			"level":      "info",
			"msg":        "http_request",
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapped.status,
			"duration_ms": duration.Milliseconds(),
			"request_id": reqID,
			"remote_addr": r.RemoteAddr,
		}

		logJSON, _ := json.Marshal(entry)
		log.Println(string(logJSON))
	})
}

// Chain applies multiple middleware to a handler in order.
//
// This is a helper that makes middleware composition more readable:
//   handler := Chain(mux, WithRequestID, WithLogging, WithTimeout(10*time.Second))
//
// Instead of:
//   handler := WithRequestID(WithLogging(WithTimeout(10*time.Second)(mux)))
//
// C# equivalent: The ASP.NET pipeline builder (app.UseX().UseY())
// Java equivalent: Spring's FilterChain / HandlerInterceptor order
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply in reverse order so the first middleware in the list is outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
