// Package main is the entry point for the shipit binary.
//
// With Cobra (Lesson 14), main() is just one line: cmd.Execute().
// All command logic lives in cmd/ subpackage files.
//
// BEFORE (manual switch/case):
//   switch os.Args[1] {
//       case "serve": ...
//       case "demo": ...
//   }
//
// AFTER (Cobra):
//   cmd.Execute()  // That's it!
//
// C# equivalent: await rootCommand.InvokeAsync(args);
// Java equivalent: new CommandLine(new App()).execute(args);
package main

import "github.com/alluri02/go-shipit/cmd/shipit/cmd"

func main() {
	cmd.Execute()
}
