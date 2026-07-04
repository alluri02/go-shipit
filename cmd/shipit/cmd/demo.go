package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/alluri02/go-shipit/internal/adapters/inmemory"
	"github.com/alluri02/go-shipit/internal/domain"
	"github.com/alluri02/go-shipit/internal/transport/worker"
	"github.com/spf13/cobra"
)

// demoCmd runs the interactive demo showing all concepts.
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run interactive demo (DI, errors, models)",
	Run: func(cmd *cobra.Command, args []string) {
		repo := inmemory.NewRepository()
		queue := inmemory.NewQueue()
		notifier := inmemory.NewNotifier()
		service := domain.NewDeployService(repo, queue, notifier)

		fmt.Println("--- Dependency Injection (Lesson 05) ---")
		env := domain.NewEnvironment("production", "eastus", "https://myapp.azurecontainerapps.io")

		deploy, err := service.StartDeploy("deploy-001", "payments-api", "v2.4.1", "github-webhook", env)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created deployment: %s (%s)\n", deploy.ID, deploy.Status)

		fetched, err := service.GetDeploy("deploy-001")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Fetched from repo: %s → %s\n", fetched.ServiceName, fetched.Environment.Name)

		fmt.Println("\n--- Error Handling (Lesson 04) ---")
		_, err = service.GetDeploy("does-not-exist")
		if err != nil {
			fmt.Printf("✓ Expected error: %v\n", err)
		}
	},
}

// processCmd runs the worker pool demo.
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Run worker processor demo (goroutines & channels)",
	RunE: func(cmd *cobra.Command, args []string) error {
		workers, _ := cmd.Flags().GetInt("workers")
		if workers <= 0 {
			workers = 3
		}

		repo := inmemory.NewRepository()
		queue := inmemory.NewQueue()
		notifier := inmemory.NewNotifier()
		service := domain.NewDeployService(repo, queue, notifier)

		fmt.Printf("--- Goroutines & Channels (Lesson 07) ---\n")
		fmt.Printf("Starting processor with %d workers...\n\n", workers)

		proc := worker.NewProcessor(service, workers, 10)
		proc.Start()

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
		time.Sleep(100 * time.Millisecond)
		proc.Stop()
		fmt.Println("\n✓ All deployments processed!")
		return nil
	},
}

func init() {
	processCmd.Flags().IntP("workers", "w", 3, "Number of concurrent workers")
	rootCmd.AddCommand(demoCmd)
	rootCmd.AddCommand(processCmd)
}
