package config_test

import (
	"os"
	"testing"

	"github.com/alluri02/go-shipit/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that might be set
	os.Unsetenv("SHIPIT_SERVER_ADDR")
	os.Unsetenv("SHIPIT_ENV")
	os.Unsetenv("SHIPIT_WORKER_COUNT")

	cfg := config.Load()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ServerAddr", cfg.ServerAddr, ":8080"},
		{"Env", cfg.Env, "development"},
		{"QueueName", cfg.QueueName, "deployments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	if cfg.WorkerCount != 3 {
		t.Errorf("WorkerCount = %d, want 3", cfg.WorkerCount)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	// Set env vars
	os.Setenv("SHIPIT_SERVER_ADDR", ":9090")
	os.Setenv("SHIPIT_ENV", "production")
	os.Setenv("SHIPIT_WORKER_COUNT", "5")
	defer func() {
		os.Unsetenv("SHIPIT_SERVER_ADDR")
		os.Unsetenv("SHIPIT_ENV")
		os.Unsetenv("SHIPIT_WORKER_COUNT")
	}()

	cfg := config.Load()

	if cfg.ServerAddr != ":9090" {
		t.Errorf("ServerAddr = %q, want %q", cfg.ServerAddr, ":9090")
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.WorkerCount != 5 {
		t.Errorf("WorkerCount = %d, want 5", cfg.WorkerCount)
	}
}

func TestConfig_Validate_Development(t *testing.T) {
	cfg := &config.Config{
		ServerAddr: ":8080",
		Env:        "development",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("development config should be valid without DB/Queue, got: %v", err)
	}
}

func TestConfig_Validate_Production_Missing(t *testing.T) {
	cfg := &config.Config{
		ServerAddr: ":8080",
		Env:        "production",
		// Missing: DatabaseURL, QueueConnectionString, SlackToken
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("production config missing required fields should fail")
	}
}

func TestConfig_Validate_Production_Complete(t *testing.T) {
	cfg := &config.Config{
		ServerAddr:            ":8080",
		Env:                   "production",
		DatabaseURL:           "mysql://user:pass@host/db",
		QueueConnectionString: "DefaultEndpointsProtocol=https;...",
		SlackToken:            "xoxb-1234567890",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("complete production config should be valid, got: %v", err)
	}
}

func TestConfig_String_MasksSecrets(t *testing.T) {
	cfg := &config.Config{
		ServerAddr:            ":8080",
		Env:                   "production",
		DatabaseURL:           "mysql://admin:supersecret@db.example.com/shipit",
		QueueConnectionString: "DefaultEndpointsProtocol=https;AccountName=foo;AccountKey=secret123",
		SlackToken:            "xoxb-1234-5678-abcdef",
		WorkerCount:           3,
		BufferSize:            100,
	}

	str := cfg.String()

	// Should NOT contain full secrets
	if contains(str, "supersecret") {
		t.Error("String() should not contain database password")
	}
	if contains(str, "secret123") {
		t.Error("String() should not contain queue key")
	}

	// Should contain partial masking
	if !contains(str, "****") {
		t.Error("String() should contain masked values")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
