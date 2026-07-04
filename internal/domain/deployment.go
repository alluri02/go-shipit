package domain

import "time"

// Deployment is the core domain model — represents one deployment job.
//
// Note: Go uses composition, not inheritance. There's no "extends" or ":".
// If you need shared behavior, you embed another struct (shown in later lessons).
//
// C# equivalent:
//   public class Deployment {
//       public string ID { get; set; }
//       public string ServiceName { get; set; }
//       ...
//   }
//
// Java equivalent:
//   public class Deployment {
//       private String id;
//       private String serviceName;
//       ...
//   }
type Deployment struct {
	ID          string
	ServiceName string
	ImageTag    string
	Environment Environment   // Embedded struct (composition, not inheritance)
	Status      DeployStatus
	RiskScore   int           // 1-10, set by AI scorer
	TriggeredBy string        // "github-webhook", "slack-chatops", "api"
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDeployment creates a Deployment in Pending status with timestamps set.
//
// Note the pointer return (*Deployment). We return a pointer because:
//   1. Deployment is a large struct — avoids copying
//   2. We want callers to mutate it (advance status, set risk score)
//
// C# equivalent: public Deployment(...) — classes are always reference types
// Java equivalent: public Deployment(...) — objects are always references
func NewDeployment(id, serviceName, imageTag, triggeredBy string, env Environment) *Deployment {
	now := time.Now()
	return &Deployment{
		ID:          id,
		ServiceName: serviceName,
		ImageTag:    imageTag,
		Environment: env,
		Status:      DeployStatusPending,
		RiskScore:   0,
		TriggeredBy: triggeredBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Advance moves the deployment to the next status.
// Uses a pointer receiver (*Deployment) because it MODIFIES the struct.
//
// Value receiver  (d Deployment)  = works on a COPY (like passing by value)
// Pointer receiver (d *Deployment) = works on the ORIGINAL (like passing by reference)
//
// C# equivalent: public void Advance(DeployStatus next) { this.Status = next; }
// Java equivalent: public void advance(DeployStatus next) { this.status = next; }
func (d *Deployment) Advance(next DeployStatus) {
	d.Status = next
	d.UpdatedAt = time.Now()
}

// IsHighRisk returns true if the AI risk scorer flagged this as risky.
// Value receiver — doesn't modify the deployment.
func (d *Deployment) IsHighRisk() bool {
	return d.RiskScore >= 7
}

// ShouldRequireApproval checks if this deployment needs human approval
// based on environment AND risk score.
func (d *Deployment) ShouldRequireApproval() bool {
	return d.Environment.RequiresApproval() || d.IsHighRisk()
}

// Duration returns how long the deployment has been running.
// Returns 0 if the deployment hasn't started yet.
func (d *Deployment) Duration() time.Duration {
	if d.Status == DeployStatusPending {
		return 0
	}
	return time.Since(d.CreatedAt)
}
