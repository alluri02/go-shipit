package domain_test

import (
	"fmt"
	"testing"

	"github.com/alluri02/go-shipit/internal/domain"
)

// TestNewDeployment verifies that NewDeployment sets all fields correctly.
//
// Go test functions:
//   - Must be in a file ending with _test.go
//   - Must start with Test
//   - Take *testing.T as the only parameter
//
// C# equivalent (xUnit):
//   [Fact]
//   public void NewDeployment_SetsFieldsCorrectly() { ... }
//
// Java equivalent (JUnit 5):
//   @Test
//   void newDeployment_setsFieldsCorrectly() { ... }
func TestNewDeployment(t *testing.T) {
	env := domain.NewEnvironment("staging", "eastus", "https://staging.app.io")
	deploy := domain.NewDeployment("d-001", "payments-api", "v2.4.1", "github-webhook", env)

	if deploy.ID != "d-001" {
		t.Errorf("ID = %q, want %q", deploy.ID, "d-001")
	}
	if deploy.ServiceName != "payments-api" {
		t.Errorf("ServiceName = %q, want %q", deploy.ServiceName, "payments-api")
	}
	if deploy.Status != domain.DeployStatusPending {
		t.Errorf("Status = %v, want %v", deploy.Status, domain.DeployStatusPending)
	}
	if deploy.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

// TestDeployment_Advance uses TABLE-DRIVEN TESTS — Go's signature testing pattern.
//
// Instead of writing many test functions, you define a table of inputs + expected outputs,
// then loop through them. This makes it easy to add new cases.
//
// C# equivalent (xUnit):
//   [Theory]
//   [InlineData(DeployStatus.Pending, DeployStatus.Building)]
//   [InlineData(DeployStatus.Building, DeployStatus.Pushing)]
//   public void Advance_SetsCorrectStatus(DeployStatus initial, DeployStatus expected) { ... }
//
// Java equivalent (JUnit 5):
//   @ParameterizedTest
//   @CsvSource({"PENDING,BUILDING", "BUILDING,PUSHING"})
//   void advance_setsCorrectStatus(DeployStatus initial, DeployStatus expected) { ... }
func TestDeployment_AdvanceSafe(t *testing.T) {
	// Table of test cases — each is a struct with a name and test data
	tests := []struct {
		name        string
		initial     domain.DeployStatus
		next        domain.DeployStatus
		wantErr     bool
	}{
		// Valid transitions
		{name: "pending to building", initial: domain.DeployStatusPending, next: domain.DeployStatusBuilding, wantErr: false},
		{name: "building to pushing", initial: domain.DeployStatusBuilding, next: domain.DeployStatusPushing, wantErr: false},
		{name: "pushing to deploying", initial: domain.DeployStatusPushing, next: domain.DeployStatusDeploying, wantErr: false},
		{name: "deploying to succeeded", initial: domain.DeployStatusDeploying, next: domain.DeployStatusSucceeded, wantErr: false},
		{name: "any to failed", initial: domain.DeployStatusBuilding, next: domain.DeployStatusFailed, wantErr: false},
		{name: "failed to rolledback", initial: domain.DeployStatusFailed, next: domain.DeployStatusRolledBack, wantErr: false},

		// Invalid transitions
		{name: "pending to succeeded", initial: domain.DeployStatusPending, next: domain.DeployStatusSucceeded, wantErr: true},
		{name: "building to pending", initial: domain.DeployStatusBuilding, next: domain.DeployStatusPending, wantErr: true},
		{name: "succeeded to building", initial: domain.DeployStatusSucceeded, next: domain.DeployStatusBuilding, wantErr: true},
	}

	for _, tt := range tests {
		// t.Run creates a subtest — like [Theory] data rows in xUnit
		t.Run(tt.name, func(t *testing.T) {
			env := domain.NewEnvironment("staging", "eastus", "")
			deploy := domain.NewDeployment("test", "svc", "v1", "api", env)
			deploy.Advance(tt.initial) // Set initial status

			err := deploy.AdvanceSafe(tt.next)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for %s → %s, got nil", tt.initial, tt.next)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for %s → %s: %v", tt.initial, tt.next, err)
			}
		})
	}
}

// TestDeployment_IsHighRisk demonstrates simple table-driven tests.
func TestDeployment_IsHighRisk(t *testing.T) {
	tests := []struct {
		score int
		want  bool
	}{
		{score: 0, want: false},
		{score: 5, want: false},
		{score: 6, want: false},
		{score: 7, want: true},
		{score: 8, want: true},
		{score: 10, want: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%d", tt.score), func(t *testing.T) {
			env := domain.NewEnvironment("staging", "eastus", "")
			deploy := domain.NewDeployment("test", "svc", "v1", "api", env)
			deploy.RiskScore = tt.score

			if got := deploy.IsHighRisk(); got != tt.want {
				t.Errorf("IsHighRisk() with score %d = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

// TestEnvironment_RequiresApproval tests environment approval logic.
func TestEnvironment_RequiresApproval(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{name: "production requires approval", env: "production", want: true},
		{name: "prod requires approval", env: "prod", want: true},
		{name: "staging does not", env: "staging", want: false},
		{name: "dev does not", env: "dev", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := domain.NewEnvironment(tt.env, "eastus", "")
			if got := env.RequiresApproval(); got != tt.want {
				t.Errorf("RequiresApproval() for %q = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}
