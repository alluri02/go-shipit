package domain

import (
	"errors"
	"fmt"
)

// Go has NO exceptions. Errors are values — returned as the last return value.
//
// C# equivalent: throw new NotFoundException("deployment not found");
// Java equivalent: throw new NotFoundException("deployment not found");
//
// In Go: return nil, ErrNotFound
//
// Key insight: In C#/Java, you throw and catch. In Go, you return and check.
// This makes error paths EXPLICIT — you always see them in the code.

// --- Sentinel Errors ---
// These are package-level variables used for comparison with errors.Is().
//
// C# equivalent: custom exception types (NotFoundException, ConflictException)
// Java equivalent: custom exception types (NotFoundException, ConflictException)
//
// Go idiom: errors start with "Err" prefix.
var (
	// ErrNotFound means the requested resource doesn't exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists means a resource with that ID already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidStatus means a state transition is not allowed.
	ErrInvalidStatus = errors.New("invalid status transition")

	// ErrUnauthorized means the caller lacks permission.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrRiskThresholdExceeded means the AI risk score is too high to proceed.
	ErrRiskThresholdExceeded = errors.New("risk threshold exceeded")
)

// --- Custom Error Types ---
// When you need more context than a simple message, define a struct that implements error.
//
// The `error` interface in Go is just:
//   type error interface {
//       Error() string
//   }
//
// Any struct with an Error() method IS an error. (Implicit interface — Lesson 03!)
//
// C# equivalent:
//   public class ValidationError : Exception {
//       public string Field { get; }
//       public ValidationError(string field, string msg) : base(msg) { Field = field; }
//   }
//
// Java equivalent:
//   public class ValidationException extends RuntimeException {
//       private final String field;
//       public ValidationException(String field, String message) { ... }
//   }

// ValidationError represents a field-level validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface — this makes ValidationError an error.
// No "implements" keyword needed (Go's implicit interfaces from Lesson 03).
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s — %s", e.Field, e.Message)
}

// DeployError wraps an error with deployment context.
// This is Go's pattern for adding context to errors — wrap them.
//
// C# equivalent:
//   public class DeployException : Exception {
//       public string DeploymentID { get; }
//       public string Stage { get; }
//       public DeployException(string id, string stage, Exception inner) : base(..., inner) { }
//   }
//
// Java equivalent:
//   public class DeployException extends RuntimeException {
//       private final String deploymentID;
//       public DeployException(String id, String stage, Throwable cause) { super(msg, cause); }
//   }
type DeployError struct {
	DeploymentID string
	Stage        string // "build", "push", "deploy", "rollback"
	Err          error  // The underlying error (for unwrapping)
}

func (e *DeployError) Error() string {
	return fmt.Sprintf("deploy %s failed at %s: %v", e.DeploymentID, e.Stage, e.Err)
}

// Unwrap returns the underlying error — enables errors.Is() and errors.As() to work.
//
// This is Go's version of InnerException (C#) / getCause() (Java).
// errors.Is(err, ErrNotFound) walks the chain via Unwrap().
func (e *DeployError) Unwrap() error {
	return e.Err
}

// --- Helper Functions for Creating Errors ---

// NewValidationError creates a validation error for a specific field.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// NewDeployError wraps an error with deployment context.
func NewDeployError(deploymentID, stage string, err error) *DeployError {
	return &DeployError{
		DeploymentID: deploymentID,
		Stage:        stage,
		Err:          err,
	}
}

// WrapNotFound wraps ErrNotFound with a formatted message.
// Uses fmt.Errorf with %w verb — Go's way to wrap errors while preserving the chain.
//
// The %w verb makes errors.Is(result, ErrNotFound) return true.
//
// C# equivalent: throw new NotFoundException($"deployment {id} not found");
// Java equivalent: throw new NotFoundException("deployment " + id + " not found");
func WrapNotFound(resource, id string) error {
	return fmt.Errorf("%s %q: %w", resource, id, ErrNotFound)
}
