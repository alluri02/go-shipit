package domain

// DeployStatus represents the state of a deployment.
//
// Go doesn't have enums. Instead, we use a custom type + iota (auto-incrementing constants).
//
// C# equivalent:
//   public enum DeployStatus { Pending, Building, Pushing, Deploying, Succeeded, Failed, RolledBack }
//
// Java equivalent:
//   public enum DeployStatus { PENDING, BUILDING, PUSHING, DEPLOYING, SUCCEEDED, FAILED, ROLLED_BACK }
type DeployStatus int

const (
	DeployStatusPending   DeployStatus = iota // 0
	DeployStatusBuilding                      // 1 (iota auto-increments)
	DeployStatusPushing                       // 2
	DeployStatusDeploying                     // 3
	DeployStatusSucceeded                     // 4
	DeployStatusFailed                        // 5
	DeployStatusRolledBack                    // 6
)

// String returns the human-readable nameirtual of the status.
// This implements the fmt.Stringer interface (implicitly — no "implements" keyword needed).
//
// C# equivalent: override string ToString()
// Java equivalent: @Override public String toString()
func (s DeployStatus) String() string {
	switch s {
	case DeployStatusPending:
		return "pending"
	case DeployStatusBuilding:
		return "building"
	case DeployStatusPushing:
		return "pushing"
	case DeployStatusDeploying:
		return "deploying"
	case DeployStatusSucceeded:
		return "succeeded"
	case DeployStatusFailed:
		return "failed"
	case DeployStatusRolledBack:
		return "rolled_back"
	default:
		return "unknown"
	}
}

// IsTerminal returns true if the deployment has reached a final state.
// This is a value receiver method — it doesn't modify the status.
//
// C# equivalent: public bool IsTerminal() => this == Succeeded || ...
// Java equivalent: public boolean isTerminal() { return this == SUCCEEDED || ... }
func (s DeployStatus) IsTerminal() bool {
	return s == DeployStatusSucceeded || s == DeployStatusFailed || s == DeployStatusRolledBack
}
