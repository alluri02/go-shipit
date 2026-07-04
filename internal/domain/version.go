// Package domain contains the core business logic of ShipIt.
// This package has ZERO external dependencies — it's pure Go.
//
// In hexagonal architecture:
//   - This is the "inside" of the hexagon
//   - No imports from adapters, no HTTP, no database
//
// C# equivalent: YourProject.Domain class library (no NuGet packages)
// Java equivalent: domain module in a multi-module Maven/Gradle project
package domain

// Version is the current version of ShipIt.
// In Go, exported identifiers start with an uppercase letter.
//
// C# equivalent: public const string Version = "0.1.0";
// Java equivalent: public static final String VERSION = "0.1.0";
const Version = "0.1.0"
