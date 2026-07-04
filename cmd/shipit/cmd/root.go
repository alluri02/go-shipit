package cmd

import (
	"fmt"
	"os"

	"github.com/alluri02/go-shipit/internal/domain"
	"github.com/spf13/cobra"
)

// rootCmd is the base command — runs when you type just `shipit`.
//
// KEY GO CONCEPT: CLI frameworks (Cobra).
// Cobra is the Go standard for CLI tools — used by kubectl, docker, gh, hugo.
// It provides: subcommands, flags, help generation, shell completion.
//
// C# equivalent: System.CommandLine (Microsoft's CLI framework)
//   var rootCommand = new RootCommand("ShipIt deployment orchestrator");
//   rootCommand.AddCommand(serveCommand);
//   rootCommand.AddCommand(demoCommand);
//   await rootCommand.InvokeAsync(args);
//
// Java equivalent: Picocli
//   @Command(name = "shipit", subcommands = {ServeCommand.class, DemoCommand.class})
//   public class App implements Runnable { ... }
var rootCmd = &cobra.Command{
	Use:   "shipit",
	Short: "ShipIt — Deployment Pipeline Orchestrator",
	Long: fmt.Sprintf(`
   _____ __    _ ____ ____ ______
  / ___// /_  (_) __ \/  _/_  __/
  \__ \/ __ \/ / /_/ // /   / /     v%s
 ___/ / / / / / ____// /   / /      Deployment Orchestrator
/____/_/ /_/_/_/   /___/  /_/       github.com/alluri02/go-shipit

ShipIt orchestrates container deployments.
Learn production-grade Go by building a real system.`, domain.Version),
}

// Execute runs the root command — called from main().
// This is the ONLY export from this package.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
