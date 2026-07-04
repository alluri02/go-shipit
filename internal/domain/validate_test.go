package domain_test

import (
	"errors"
	"testing"

	"github.com/alluri02/go-shipit/internal/domain"
)

// TestValidateDeployment uses table-driven tests for validation.
func TestValidateDeployment(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		imageTag    string
		triggeredBy string
		wantErr     bool
		wantField   string // expected field in ValidationError
	}{
		{
			name:        "valid input",
			serviceName: "payments-api",
			imageTag:    "v2.4.1",
			triggeredBy: "api",
			wantErr:     false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			imageTag:    "v1.0",
			triggeredBy: "api",
			wantErr:     true,
			wantField:   "serviceName",
		},
		{
			name:        "empty image tag",
			serviceName: "my-service",
			imageTag:    "",
			triggeredBy: "api",
			wantErr:     true,
			wantField:   "imageTag",
		},
		{
			name:        "empty triggered by",
			serviceName: "my-service",
			imageTag:    "v1.0",
			triggeredBy: "",
			wantErr:     true,
			wantField:   "triggeredBy",
		},
		{
			name:        "image tag too long",
			serviceName: "my-service",
			imageTag:    string(make([]byte, 129)), // 129 chars
			triggeredBy: "api",
			wantErr:     true,
			wantField:   "imageTag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateDeployment(tt.serviceName, tt.imageTag, tt.triggeredBy)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check specific field if expected
			if tt.wantErr && tt.wantField != "" {
				var valErr *domain.ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if valErr.Field != tt.wantField {
					t.Errorf("error field = %q, want %q", valErr.Field, tt.wantField)
				}
			}
		})
	}
}
