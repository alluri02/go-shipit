# Lesson 01: Go Modules & Packages

## What We Built
```
go-shipit/
├── go.mod                      ← Module definition (like .csproj or pom.xml)
├── cmd/
│   └── shipit/
│       └── main.go             ← Entry point (package main)
└── internal/
    └── domain/
        └── version.go          ← Domain package (business logic)
```

---

## Concept Map: Go vs C# vs Java

| Concept | Go | C# | Java |
|---------|-----|-----|------|
| **Project file** | `go.mod` | `.csproj` / `.sln` | `pom.xml` / `build.gradle` |
| **Module name** | `module github.com/alluri02/go-shipit` | `<RootNamespace>` in .csproj | `groupId + artifactId` |
| **Package** | `package domain` (folder = package) | `namespace ShipIt.Domain` | `package com.shipit.domain` |
| **Entry point** | `func main()` in `package main` | `static void Main()` in Program.cs | `public static void main(String[] args)` |
| **Visibility** | Uppercase = public, lowercase = private | `public` / `private` keywords | `public` / `private` keywords |
| **Internal code** | `internal/` folder (compiler-enforced) | `internal` access modifier | Package-private (no modifier) |
| **Dependency mgmt** | `go get` + `go.sum` | NuGet + `packages.lock.json` | Maven Central + dependency lock |

---

## Deep Dive: Go Modules

### What is `go.mod`?

```go
module github.com/alluri02/go-shipit

go 1.26.4
```

This file declares:
1. **Module path** — the import path other packages use to reference your code
2. **Go version** — minimum Go version required

### C# Equivalent (.csproj)
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <RootNamespace>GoShipIt</RootNamespace>
  </PropertyGroup>
</Project>
```

### Java Equivalent (pom.xml)
```xml
<project>
  <groupId>com.github.alluri02</groupId>
  <artifactId>go-shipit</artifactId>
  <version>0.1.0</version>
</project>
```

---

## Deep Dive: Packages

### Rule: One folder = One package

In Go, **every `.go` file in a folder must declare the same package name**. The folder name IS the package.

```
internal/domain/         ← all files here say: package domain
internal/domain/version.go
internal/domain/deploy.go   (future)
internal/domain/build.go    (future)
```

### C# Comparison
In C#, you can have multiple namespaces per file and multiple files per namespace. Go is stricter — one package per directory, enforced by the compiler.

```csharp
// C# — namespace can be anything, independent of folder
namespace ShipIt.Domain
{
    public static class AppInfo
    {
        public const string Version = "0.1.0";
    }
}
```

### Java Comparison
Java's package maps to folder structure too, but requires explicit `package` declaration matching the directory path:

```java
// Java — must match folder: src/main/java/com/shipit/domain/AppInfo.java
package com.shipit.domain;

public class AppInfo {
    public static final String VERSION = "0.1.0";
}
```

---

## Deep Dive: Visibility (Exported vs Unexported)

Go has **no keywords** for visibility. It uses **capitalization**:

```go
package domain

const Version = "0.1.0"   // Uppercase V → exported (public)
const version = "0.1.0"   // Lowercase v → unexported (private to package)
```

### Comparison Table

| Go | C# | Java |
|----|-----|------|
| `func DoSomething()` | `public void DoSomething()` | `public void doSomething()` |
| `func doSomething()` | `private void DoSomething()` | `private void doSomething()` |
| `type Service struct` | `public class Service` | `public class Service` |
| `type service struct` | `internal class Service` | `class Service` (package-private) |

### Why?
Go's philosophy: less ceremony. You can tell at a glance whether something is public by its first letter. No need to scan for access modifiers.

---

## Deep Dive: The `internal/` Directory

The `internal/` directory is **compiler-enforced** in Go. Code inside `internal/` can only be imported by code in the parent directory tree.

```
go-shipit/
├── cmd/shipit/main.go          ← CAN import internal/domain ✓
├── internal/
│   └── domain/version.go       ← Protected
└── pkg/                        ← (future) Public library code anyone can import
```

If someone does `go get github.com/alluri02/go-shipit` and tries to import `internal/domain`, the compiler **refuses**.

### C# Equivalent
```csharp
// C# uses the `internal` access modifier
internal class DeployService { }  // Only visible within the same assembly
```

### Java Equivalent
Java doesn't have a direct equivalent. The closest is:
- Module system (`module-info.java`) with unexported packages
- Or simply package-private access (no modifier)

---

## Deep Dive: Imports

```go
import (
    "fmt"                                        // Standard library
    "os"                                         // Standard library
    "github.com/alluri02/go-shipit/internal/domain"  // Our package
)
```

### Key differences from C#/Java:

| Aspect | Go | C# | Java |
|--------|-----|-----|------|
| Import unit | Package (directory) | Namespace or type | Package or class |
| Unused import | **Compile error** | Warning | Warning |
| Aliasing | `alias "path/pkg"` | `using Alias = Namespace` | Not possible (except static) |
| Wildcard | Not possible | `using Namespace;` (all types) | `import package.*;` |

### Go is strict: unused imports = compile error

```go
import "fmt"  // If you don't use fmt, your code WON'T COMPILE

// C# — unused `using` is just a warning
// Java — unused import is just a warning
```

This keeps Go codebases clean. Your editor (VS Code + Go extension) auto-removes unused imports on save.

---

## Try It

```bash
# Build and run
cd cmd/shipit
go run . httpservice

# Output: ShipIt v0.1.0 — starting httpservice
```

### What `go run` does:
1. Compiles all `.go` files in the current package
2. Links them into a temporary binary
3. Executes the binary

### C# Equivalent: `dotnet run -- httpservice`
### Java Equivalent: `mvn exec:java -Dexec.args="httpservice"`

---

## Key Takeaways

1. **`go.mod`** = your project file. One per module. Tracks dependencies.
2. **Folder = Package**. No exceptions. Keep it simple.
3. **Uppercase = Public**. No keywords needed.
4. **`internal/`** = compiler-enforced encapsulation (like `internal` in C#).
5. **Unused imports** = compile error. Go forces you to keep code clean.
6. **`cmd/`** pattern = standard Go layout for binaries.

---

## Next: [Lesson 02 — Types, Structs & Methods](./02-types-structs-methods.md)
We'll define the `Deployment` and `Environment` domain models using structs and methods.
