// Package domain — Lesson 04 addition: ValidateDeployment demonstrates error handling.
package domain

import "fmt"

// ValidateDeployment checks if a deployment request is valid.
// Returns nil if valid, or a *ValidationError if not.
//
// This shows Go's pattern: validate and return errors, don't throw exceptions.
//
// C# equivalent:
//   public void Validate() {
//       if (string.IsNullOrEmpty(ServiceName))
//           throw new ValidationException("serviceName", "must not be empty");
//   }
//
// Java equivalent:
//   public void validate() {
//       if (serviceName == null || serviceName.isEmpty())
//           throw new ValidationException("serviceName", "must not be empty");
//   }
func ValidateDeployment(serviceName, imageTag, triggeredBy string) error {
	if serviceName == "" {
		return NewValidationError("serviceName", "must not be empty")
	}
	if imageTag == "" {
		return NewValidationError("imageTag", "must not be empty")
	}
	if triggeredBy == "" {
		return NewValidationError("triggeredBy", "must not be empty")
	}

	// Validate imageTag format (must contain a colon or be a valid tag)
	if len(imageTag) > 128 {
		return NewValidationError("imageTag", "must be 128 characters or fewer")
	}

	return nil // nil means "no error" — success!
}

// AdvanceSafe is like Advance but validates the state transition first.
// Returns ErrInvalidStatus if the transition isn't allowed.
//
// This demonstrates multiple return values + error handling:
//   result, err := something()
//   if err != nil { ... }
func (d *Deployment) AdvanceSafe(next DeployStatus) error {
	if !d.canTransitionTo(next) {
		return fmt.Errorf(
			"cannot transition from %s to %s: %w",
			d.Status, next, ErrInvalidStatus,
		)
	}
	d.Advance(next)
	return nil
}

// canTransitionTo defines allowed state transitions.
func (d *Deployment) canTransitionTo(next DeployStatus) bool {
	allowed := map[DeployStatus][]DeployStatus{
		DeployStatusPending:   {DeployStatusBuilding, DeployStatusFailed},
		DeployStatusBuilding:  {DeployStatusPushing, DeployStatusFailed},
		DeployStatusPushing:   {DeployStatusDeploying, DeployStatusFailed},
		DeployStatusDeploying: {DeployStatusSucceeded, DeployStatusFailed},
		DeployStatusFailed:    {DeployStatusRolledBack},
	}

	targets, ok := allowed[d.Status]
	if !ok {
		return false
	}

	for _, t := range targets {
		if t == next {
			return true
		}
	}
	return false
}
