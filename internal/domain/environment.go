package domain

// Environment represents a deployment target (e.g., staging, production).
//
// Go structs are VALUE types by default (like C# structs).
// They're allocated on the stack unless you take their address (&).
//
// C# equivalent:
//   public class Environment { public string Name { get; set; } ... }
//
// Java equivalent:
//   public class Environment { private String name; ... }
type Environment struct {
	Name       string // "staging", "production" — exported (public) field
	Region     string // "eastus", "westeurope"
	ClusterURL string // Azure Container Apps endpoint
	IsProduction bool
}

// NewEnvironment creates an Environment with validation.
// Go has no constructors — we use "New*" factory functions by convention.
//
// C# equivalent:
//   public Environment(string name, string region, string clusterURL) { ... }
//
// Java equivalent:
//   public Environment(String name, String region, String clusterURL) { ... }
func NewEnvironment(name, region, clusterURL string) Environment {
	return Environment{
		Name:         name,
		Region:       region,
		ClusterURL:   clusterURL,
		IsProduction: name == "production" || name == "prod",
	}
}

// RequiresApproval returns true if deployments to this environment need manual approval.
// This is a method with a value receiver (env Environment).
//
// C# equivalent: public bool RequiresApproval() => IsProduction;
// Java equivalent: public boolean requiresApproval() { return isProduction; }
func (env Environment) RequiresApproval() bool {
	return env.IsProduction
}
