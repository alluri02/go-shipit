# Lesson 14: CLI with Cobra

## What We Built
```
go-shipit/
└── cmd/
    └── shipit/
        ├── main.go          ← 3 lines! Just calls cmd.Execute()
        └── cmd/
            ├── root.go      ← Root command + banner
            ├── version.go   ← version subcommand
            ├── serve.go     ← serve subcommand (with flags)
            └── demo.go      ← demo + process subcommands
```

---

## The Transformation

### BEFORE (manual switch/case):
```go
func main() {
    command := os.Args[1]
    switch command {
    case "serve": ...
    case "demo": ...
    case "version": ...
    default: fmt.Println("unknown command")
    }
}
```

### AFTER (Cobra):
```go
func main() {
    cmd.Execute()  // That's it!
}
```

---

## Why Cobra?

Cobra is THE CLI framework in Go. Used by: **kubectl**, **docker**, **gh** (GitHub CLI), **hugo**, **terraform**.

| Feature | Manual (before) | Cobra (after) |
|---------|----------------|---------------|
| Help text | Hand-written | Auto-generated |
| Flags | Manual `os.Args` parsing | `--port`, `-p` with types |
| Subcommands | switch/case | Tree structure |
| Shell completion | Not possible | Built-in (bash/zsh/fish/powershell) |
| Validation | Manual | Built-in (required flags, arg count) |
| Error handling | Manual | Automatic |

---

## Concept Map

| | Go (Cobra) | C# (System.CommandLine) | Java (Picocli) |
|-|------------|------------------------|----------------|
| **Framework** | `github.com/spf13/cobra` | `System.CommandLine` | `picocli` |
| **Root command** | `cobra.Command{}` | `new RootCommand()` | `@Command(name="app")` |
| **Subcommand** | `rootCmd.AddCommand(sub)` | `root.AddCommand(sub)` | `subcommands={Sub.class}` |
| **Flags** | `cmd.Flags().StringP(...)` | `new Option<string>("--port")` | `@Option(names="--port")` |
| **Execute** | `rootCmd.Execute()` | `rootCommand.InvokeAsync(args)` | `new CommandLine(app).execute(args)` |
| **Help** | Automatic | Automatic | Automatic |
| **Completion** | Built-in | Manual | Built-in |

---

## Pattern 1: Defining a Command

```go
var serveCmd = &cobra.Command{
    Use:   "serve",              // Command name
    Short: "Start the HTTP API", // One-line description (shown in help list)
    Long:  "Full description",   // Shown in `shipit serve --help`
    RunE: func(cmd *cobra.Command, args []string) error {
        // Command logic here
        return nil  // or return err
    },
}
```

### C# Equivalent:
```csharp
var serveCommand = new Command("serve", "Start the HTTP API");
serveCommand.SetHandler((InvocationContext ctx) => {
    // Command logic
    return Task.FromResult(0);
});
```

### Java Equivalent:
```java
@Command(name = "serve", description = "Start the HTTP API")
public class ServeCommand implements Callable<Integer> {
    @Override
    public Integer call() {
        // Command logic
        return 0;
    }
}
```

---

## Pattern 2: Flags (Typed Parameters)

```go
// Define flag
serveCmd.Flags().StringP("port", "p", "", "Listen port")
serveCmd.Flags().IntP("workers", "w", 0, "Worker count")

// Read flag in RunE
port, _ := cmd.Flags().GetString("port")
workers, _ := cmd.Flags().GetInt("workers")
```

Usage: `shipit serve --port 9090 -w 5`

### C# Equivalent:
```csharp
var portOption = new Option<string>("--port", "Listen port");
portOption.AddAlias("-p");
serveCommand.AddOption(portOption);
```

### Java Equivalent:
```java
@Option(names = {"-p", "--port"}, description = "Listen port")
private String port;

@Option(names = {"-w", "--workers"}, description = "Worker count")
private int workers;
```

---

## Pattern 3: init() for Registration

```go
func init() {
    rootCmd.AddCommand(serveCmd)
}
```

Go's `init()` functions run automatically when the package is imported. Each file in `cmd/` registers its command in `init()`.

### Why init()?
- Commands register themselves — no central "registry" file
- Adding a new command = add a new file. Done.
- No need to edit existing files

### C# Equivalent:
```csharp
// In Program.cs — must manually add each command
rootCommand.AddCommand(serveCommand);
rootCommand.AddCommand(demoCommand);
```

### Java Equivalent:
```java
// In annotation — must list all subcommands
@Command(subcommands = {ServeCommand.class, DemoCommand.class})
```

---

## Pattern 4: RunE vs Run

| Method | Return | Use when |
|--------|--------|----------|
| `Run` | `func(cmd, args)` | Command never fails |
| `RunE` | `func(cmd, args) error` | Command can fail (I/O, config, etc.) |

```go
// RunE — returns error, Cobra prints it and exits with code 1
RunE: func(cmd *cobra.Command, args []string) error {
    if err := startServer(); err != nil {
        return err  // Cobra prints: "Error: <message>"
    }
    return nil
},
```

---

## Pattern 5: Shell Completion (Free!)

Cobra generates shell completion scripts automatically:

```bash
# PowerShell
shipit completion powershell | Out-String | Invoke-Expression

# Bash
source <(shipit completion bash)

# Zsh
shipit completion zsh > ~/.zsh/completions/_shipit
```

Now `shipit <TAB>` shows: `serve demo process version completion help`

---

## Project Structure After Cobra

```
cmd/shipit/
├── main.go          ← 3 lines (package main, import, cmd.Execute())
└── cmd/
    ├── root.go      ← Root command definition
    ├── serve.go     ← Each file = one subcommand
    ├── version.go
    └── demo.go
```

### Adding a new command:
1. Create `cmd/shipit/cmd/newcommand.go`
2. Define `var newCmd = &cobra.Command{...}`
3. Add `func init() { rootCmd.AddCommand(newCmd) }`

That's it. No other files to edit.

---

## Try It

```bash
go build -o shipit.exe ./cmd/shipit

# Auto-generated help
.\shipit.exe --help
.\shipit.exe serve --help

# Flags
.\shipit.exe serve --port 9090 --workers 5

# Shell completion
.\shipit.exe completion powershell

# All existing commands still work
.\shipit.exe version
.\shipit.exe demo
.\shipit.exe process --workers 5
```

---

## Key Takeaways

1. **Cobra** = the standard Go CLI framework (kubectl, docker, gh all use it).
2. **`main.go` becomes 3 lines** — all logic moves to `cmd/` package files.
3. **Each file = one command** — `init()` registers it automatically.
4. **Flags are typed** — `StringP`, `IntP`, `BoolP` with short aliases.
5. **Help + completion are free** — auto-generated from your command definitions.
6. **`RunE` for fallible commands** — return `error`, Cobra handles display + exit code.

---

## Next: [Lesson 15 — CI/CD with GitHub Actions](./15-cicd.md)
We'll create a production CI/CD pipeline: lint → test → build → push.
