// Package main is the entry point for the shipit binary.
// In Go, every executable must have a `main` package with a `main()` function.
//
// C# equivalent: Program.cs with static void Main(string[] args)
// Java equivalent: public static void main(String[] args)
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alluri02/go-shipit/internal/adapters/inmemory"
	"github.com/alluri02/go-shipit/internal/config"
	"github.com/alluri02/go-shipit/internal/domain"
	httpserver "github.com/alluri02/go-shipit/internal/transport/http"
	"github.com/alluri02/go-shipit/internal/transport/worker"
)

func main() {
	// os.Args gives command-line arguments (like args[] in C#/Java)
	if len(os.Args) < 2 {
		printBanner()
		fmt.Println("\nUsage: shipit <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  serve          Start all services locally")
		fmt.Println("  demo           Run demo (DI + error handling)")
		fmt.Println("  process        Run worker processor demo (goroutines)")
		fmt.Println("  version        Print version info")
		fmt.Println("  help           Show this message")
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Printf("shipit v%s\n", domain.Version)
	case "serve":
		// --- Lesson 09: Configuration ---
		// Load config from environment variables (12-Factor App pattern).
		cfg := config.Load()
		if err := cfg.Validate(); err != nil {
			log.Fatalf("config error: %v", err)
		}
		log.Printf("Config: %s", cfg)

		// Wire dependencies (Lesson 05) then start the HTTP server.
		repo := inmemory.NewRepository()
		queue := inmemory.NewQueue()
		notifier := inmemory.NewNotifier()
		service := domain.NewDeployService(repo, queue, notifier)

		// Create and start the server using config
		srv := httpserver.NewServer(cfg.ServerAddr, service)
		printBanner()
		fmt.Printf("HTTP API running at http://localhost%s\n", cfg.ServerAddr)
		fmt.Printf("Environment: %s | Workers: %d\n", cfg.Env, cfg.WorkerCount)
		fmt.Println("Endpoints:")
		fmt.Println("  GET  /health")
		fmt.Println("  POST /deploys")
		fmt.Println("  GET  /deploys/{id}")
		fmt.Println("  GET  /deploys?service=name")
		fmt.Print("\nPress Ctrl+C to stop.\n\n")
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
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
	case "process":
		// --- Lesson 07: Goroutines & Channels ---
		// Demonstrates concurrent processing with worker pool pattern.
		repo := inmemory.NewRepository()
		queue := inmemory.NewQueue()
		notifier := inmemory.NewNotifier()
		service := domain.NewDeployService(repo, queue, notifier)

		fmt.Println("--- Goroutines & Channels (Lesson 07) ---")
		fmt.Println("Starting processor with 3 workers...")
		fmt.Println()

		// Create processor: 3 workers, buffer of 10 jobs
		proc := worker.NewProcessor(service, 3, 10)
		proc.Start()

		// Submit 5 jobs — they'll be processed concurrently by 3 workers
		jobs := []worker.DeployJob{
			{DeploymentID: "deploy-001", ServiceName: "payments-api", ImageTag: "v2.4.1", Environment: "staging", Region: "eastus"},
			{DeploymentID: "deploy-002", ServiceName: "auth-service", ImageTag: "v1.0.0", Environment: "staging", Region: "eastus"},
			{DeploymentID: "deploy-003", ServiceName: "notifications", ImageTag: "v3.2.0", Environment: "production", Region: "westus"},
			{DeploymentID: "deploy-004", ServiceName: "payments-api", ImageTag: "v2.4.2", Environment: "production", Region: "eastus"},
			{DeploymentID: "deploy-005", ServiceName: "gateway", ImageTag: "v5.1.0", Environment: "staging", Region: "westeurope"},
		}

		for _, job := range jobs {
			proc.Submit(job)
			fmt.Printf("→ Submitted: %s (%s)\n", job.DeploymentID, job.ServiceName)
		}

		fmt.Print("\nAll jobs submitted. Workers processing concurrently...\n\n")
		time.Sleep(100 * time.Millisecond) // Let workers start before we close

		// Stop and wait for all workers to complete
		proc.Stop()
		fmt.Println("\n✓ All deployments processed!")
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
