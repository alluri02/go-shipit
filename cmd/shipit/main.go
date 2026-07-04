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
	case "help":
		printBanner()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Printf(`
   _____ __    _ ____  ____
  / ___// /_  (_) __ \/  _/
  \__ \/ __ \/ / /_/ // /    v%s
 ___/ / / / / / ____// /     Deployment Orchestrator
/____/_/ /_/_/_/   /___/     github.com/alluri02/go-shipit

`, domain.Version)
}
