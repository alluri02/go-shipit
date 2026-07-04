package cmd

import (
	"fmt"

	"github.com/alluri02/go-shipit/internal/domain"
	"github.com/spf13/cobra"
)

// versionCmd prints the version — like `kubectl version` or `docker --version`.
//
// C# equivalent:
//   var versionCommand = new Command("version", "Print version info");
//   versionCommand.SetHandler(() => Console.WriteLine($"shipit v{Version}"));
//
// Java equivalent (Picocli):
//   @Command(name = "version")
//   public class VersionCommand implements Runnable {
//       public void run() { System.out.println("shipit v" + VERSION); }
//   }
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("shipit v%s\n", domain.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
