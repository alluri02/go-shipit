package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration.
//
// KEY GO CONCEPT: Configuration via environment variables (12-Factor App).
// Go has no built-in config framework — you read env vars with os.Getenv().
// For complex apps, libraries like "envconfig" or "viper" exist, but stdlib is enough.
//
// C# equivalent:
//   public class AppSettings {
//       public string ServerAddr { get; set; }
//       public string DatabaseURL { get; set; }
//   }
//   // Loaded via: builder.Configuration.GetSection("App").Bind(settings);
//   // Or: builder.Configuration["App:ServerAddr"]
//
// Java equivalent:
//   @ConfigurationProperties(prefix = "app")
//   public class AppConfig {
//       private String serverAddr;
//       private String databaseUrl;
//   }
//   // Loaded via: application.yml or application.properties
type Config struct {
	// Server
	ServerAddr string // HTTP server listen address (e.g., ":8080")

	// Database
	DatabaseURL string // MySQL connection string

	// Queue
	QueueConnectionString string // Azure Queue Storage connection string
	QueueName             string // Queue name for deploy jobs

	// Slack
	SlackToken     string // Slack bot token
	SlackChannelID string // Default notification channel

	// Worker
	WorkerCount  int // Number of concurrent processor workers
	BufferSize   int // Job channel buffer size

	// Environment
	Env string // "development", "staging", "production"
}

// Load reads configuration from environment variables with sensible defaults.
//
// Go pattern: read from env vars, fall back to defaults.
// No YAML, no JSON config files — environment variables are the standard.
// (12-Factor App: https://12factor.net/config)
//
// C# equivalent:
//   var config = new ConfigurationBuilder()
//       .AddEnvironmentVariables()
//       .AddJsonFile("appsettings.json")
//       .Build();
//
// Java equivalent:
//   # application.yml
//   app:
//     server-addr: ${SERVER_ADDR:localhost:8080}
//     database-url: ${DATABASE_URL:}
func Load() *Config {
	return &Config{
		ServerAddr:            getEnv("SHIPIT_SERVER_ADDR", ":8080"),
		DatabaseURL:           getEnv("SHIPIT_DATABASE_URL", ""),
		QueueConnectionString: getEnv("SHIPIT_QUEUE_CONNECTION_STRING", ""),
		QueueName:             getEnv("SHIPIT_QUEUE_NAME", "deployments"),
		SlackToken:            getEnv("SHIPIT_SLACK_TOKEN", ""),
		SlackChannelID:        getEnv("SHIPIT_SLACK_CHANNEL_ID", ""),
		WorkerCount:           getEnvInt("SHIPIT_WORKER_COUNT", 3),
		BufferSize:            getEnvInt("SHIPIT_BUFFER_SIZE", 100),
		Env:                   getEnv("SHIPIT_ENV", "development"),
	}
}

// Validate checks that required config is present for the given mode.
// Returns an error listing all missing values (not just the first one).
func (c *Config) Validate() error {
	var missing []string

	if c.ServerAddr == "" {
		missing = append(missing, "SHIPIT_SERVER_ADDR")
	}

	// In production, database and queue are required
	if c.IsProduction() {
		if c.DatabaseURL == "" {
			missing = append(missing, "SHIPIT_DATABASE_URL")
		}
		if c.QueueConnectionString == "" {
			missing = append(missing, "SHIPIT_QUEUE_CONNECTION_STRING")
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

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production" || c.Env == "prod"
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development" || c.Env == "dev" || c.Env == ""
}

// String returns a safe representation (no secrets) for logging.
//
// IMPORTANT: Never log secrets. Mask tokens, passwords, connection strings.
//
// C# equivalent: override ToString() with masked fields
// Java equivalent: @Override toString() with masked fields
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{env=%s, addr=%s, workers=%d, buffer=%d, db=%s, queue=%s, slack=%s}",
		c.Env,
		c.ServerAddr,
		c.WorkerCount,
		c.BufferSize,
		maskSecret(c.DatabaseURL),
		maskSecret(c.QueueConnectionString),
		maskSecret(c.SlackToken),
	)
}

// --- Helper functions ---

// getEnv reads an environment variable or returns a default.
//
// C# equivalent: Environment.GetEnvironmentVariable("KEY") ?? "default"
// Java equivalent: System.getenv().getOrDefault("KEY", "default")
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt reads an environment variable as an integer or returns a default.
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return n
}

// maskSecret hides secret values for safe logging.
func maskSecret(s string) string {
	if s == "" {
		return "<not set>"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
