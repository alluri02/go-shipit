// Package main is the entry point for the shipit binary.
// In Go, every executable must have a `main` package with a `main()` function.
//
// C# equivalent: Program.cs with static void Main(string[] args)
// Java equivalent: public static void main(String[] args)
package main

import (
	"fmt"
	"os"

	"github.com/alluri02/go-shipit/internal/domain"
)

func main() {
	// os.Args gives command-line arguments (like args[] in C#/Java)
	if len(os.Args) < 2 {
		printBanner()
		fmt.Println("\nUsage: shipit <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  serve          Start all services locally")
		fmt.Println("  version        Print version info")
		fmt.Println("  help           Show this message")
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Printf("shipit v%s\n", domain.Version)
	case "serve":
		fmt.Printf("🚀 ShipIt v%s — starting local dev server...\n", domain.Version)
		fmt.Println("   HTTP API:    http://localhost:8080")
		fmt.Println("   Webhooks:    http://localhost:8081")
		fmt.Println("   ChatOps:     connected")
		fmt.Println("   Processor:   2 workers")
	case "demo":
		// Demonstrates domain models from Lesson 02
		env := domain.NewEnvironment("production", "eastus", "https://myapp.azurecontainerapps.io")
		deploy := domain.NewDeployment("deploy-001", "payments-api", "v2.4.1", "github-webhook", env)

		fmt.Printf("Deployment: %s → %s (%s)\n", deploy.ServiceName, deploy.Environment.Name, deploy.Status)
		fmt.Printf("Requires approval: %v\n", deploy.ShouldRequireApproval())

		deploy.Advance(domain.DeployStatusBuilding)
		fmt.Printf("Status after advance: %s\n", deploy.Status)

		deploy.RiskScore = 8
		fmt.Printf("High risk: %v (score: %d)\n", deploy.IsHighRisk(), deploy.RiskScore)

		// Lesson 04: Error handling demo
		fmt.Println("\n--- Error Handling (Lesson 04) ---")

		// Validation errors
		if err := domain.ValidateDeployment("", "v1.0", "api"); err != nil {
			fmt.Printf("Validation error: %v\n", err)
		}

		// State transition errors (AdvanceSafe)
		if err := deploy.AdvanceSafe(domain.DeployStatusPending); err != nil {
			fmt.Printf("State error: %v\n", err)
		}

		// Wrapping errors
		wrapped := domain.WrapNotFound("deployment", "deploy-999")
		fmt.Printf("Wrapped error: %v\n", wrapped)
	case "help":
		printBanner()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Printf(`
   _____ __    _ ____ ____ ______
  / ___// /_  (_) __ \/  _/_  __/
  \__ \/ __ \/ / /_/ // /   / /     v%s
 ___/ / / / / / ____// /   / /      Deployment Orchestrator
/____/_/ /_/_/_/   /___/  /_/       github.com/alluri02/go-shipit

`, domain.Version)
}
