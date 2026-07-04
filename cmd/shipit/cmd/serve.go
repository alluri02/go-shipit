package cmd

import (
	"fmt"
	"log"

	"github.com/alluri02/go-shipit/internal/adapters/inmemory"
	"github.com/alluri02/go-shipit/internal/config"
	"github.com/alluri02/go-shipit/internal/domain"
	httpserver "github.com/alluri02/go-shipit/internal/transport/http"
	"github.com/spf13/cobra"
)

// serveCmd starts the HTTP server.
//
// Flags:
//   --port, -p : override the listen port
//   --workers, -w : override worker count
//
// C# equivalent:
//   var serveCommand = new Command("serve", "Start the HTTP server");
//   serveCommand.AddOption(new Option<int>("--port", getDefaultValue: () => 8080));
//
// Java equivalent (Picocli):
//   @Command(name = "serve")
//   public class ServeCommand implements Runnable {
//       @Option(names = {"-p", "--port"}, defaultValue = "8080")
//       private int port;
//   }
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long:  "Start the ShipIt HTTP server with graceful shutdown support.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config (Lesson 09)
		cfg := config.Load()

		// Override with flags if provided
		if port, _ := cmd.Flags().GetString("port"); port != "" {
			cfg.ServerAddr = ":" + port
		}
		if workers, _ := cmd.Flags().GetInt("workers"); workers > 0 {
			cfg.WorkerCount = workers
		}

		// Validate
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("config error: %v", err)
		}
		log.Printf("Config: %s", cfg)

		// Wire dependencies (Lesson 05)
		repo := inmemory.NewRepository()
		queue := inmemory.NewQueue()
		notifier := inmemory.NewNotifier()
		service := domain.NewDeployService(repo, queue, notifier)

		// Start with graceful shutdown (Lesson 11)
		srv := httpserver.NewGracefulServer(cfg.ServerAddr, service)
		fmt.Printf("HTTP API running at http://localhost%s\n", cfg.ServerAddr)
		fmt.Printf("Environment: %s | Workers: %d\n", cfg.Env, cfg.WorkerCount)
		fmt.Println("Press Ctrl+C to stop.")
		return srv.StartWithGracefulShutdown()
	},
}

func init() {
	// Define flags for this command
	//
	// C# equivalent: new Option<string>("--port", "Listen port")
	// Java equivalent: @Option(names = "--port", description = "Listen port")
	serveCmd.Flags().StringP("port", "p", "", "Listen port (overrides SHIPIT_SERVER_ADDR)")
	serveCmd.Flags().IntP("workers", "w", 0, "Worker count (overrides SHIPIT_WORKER_COUNT)")

	rootCmd.AddCommand(serveCmd)
}
