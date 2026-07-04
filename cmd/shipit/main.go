// Package main is the entry point for the shipit binary.
// In Go, every executable must have a `main` package with a `main()` function.
//
// C# equivalent: Program.cs with static void Main(string[] args)
// Java equivalent: public static void main(String[] args)
package main

import (
	"fmt"
	"os"

	"github.com/alluri02/go-shipit/internal/adapters/inmemory"
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
		// --- Lesson 05: Dependency Injection ---
		// This is the "composition root" — where we wire all dependencies.
		// In C#, this is Startup.cs / Program.cs with services.AddScoped<>().
		// In Java, this is Spring's @Configuration class.
		// In Go, it's just... function calls in main().

		// Step 1: Create adapters (concrete implementations)
		repo := inmemory.NewRepository()
		queue := inmemory.NewQueue()
		notifier := inmemory.NewNotifier()

		// Step 2: Inject into domain service (constructor injection)
		service := domain.NewDeployService(repo, queue, notifier)

		// Step 3: Use the service — it doesn't know about inmemory adapters
		fmt.Println("--- Dependency Injection (Lesson 05) ---")
		env := domain.NewEnvironment("production", "eastus", "https://myapp.azurecontainerapps.io")

		deploy, err := service.StartDeploy("deploy-001", "payments-api", "v2.4.1", "github-webhook", env)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created deployment: %s (%s)\n", deploy.ID, deploy.Status)

		// Step 4: Retrieve it back from the repo (through the service)
		fetched, err := service.GetDeploy("deploy-001")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Fetched from repo: %s → %s\n", fetched.ServiceName, fetched.Environment.Name)

		// Step 5: Show error handling — try to fetch a non-existent deploy
		fmt.Println("\n--- Error Handling (Lesson 04) ---")
		_, err = service.GetDeploy("does-not-exist")
		if err != nil {
			fmt.Printf("✓ Expected error: %v\n", err)
		}
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
