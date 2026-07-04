# Lesson 15: CI/CD with GitHub Actions

## What We Built
```
go-shipit/
├── .github/
│   └── workflows/
│       └── ci.yml          ← CI pipeline (lint → test → build → docker)
└── Dockerfile              ← Multi-stage build (10MB final image)
```

---

## Pipeline Overview

```
PR opened / push to main
        ↓
┌───────────────────────────────────────────────────┐
│  lint        │  test                              │  ← Run in parallel
│  golangci    │  go test -race -cover             │
└──────┬───────┴──────────┬────────────────────────┘
       │                  │
       └────────┬─────────┘
                ↓
┌───────────────────────────────────────────────────┐
│  build (matrix: linux/amd64, linux/arm64,         │
│         windows/amd64, darwin/arm64)              │
└──────────────────────────┬────────────────────────┘
                           ↓ (main branch only)
┌───────────────────────────────────────────────────┐
│  docker: build → push to ghcr.io                  │
└───────────────────────────────────────────────────┘
```

---

## Comparison: CI/CD Across Languages

| | Go | C# | Java |
|-|-----|-----|------|
| **Lint** | `golangci-lint` | `dotnet format --verify` + Roslyn | Checkstyle + SpotBugs |
| **Test** | `go test -race ./...` | `dotnet test` | `mvn test` |
| **Coverage** | `-coverprofile` (built-in) | Coverlet | JaCoCo |
| **Build** | `go build` (single static binary) | `dotnet publish` | `mvn package` (JAR) |
| **Cross-compile** | `GOOS=linux GOARCH=arm64` (trivial!) | `-r linux-arm64` | GraalVM native-image (complex) |
| **Docker size** | ~10MB (FROM scratch) | ~100MB (aspnet-alpine) | ~150MB (jre-alpine) |
| **Race detector** | `-race` flag (built-in!) | No equivalent | No equivalent |

---

## Deep Dive: Go's Race Detector

```yaml
- run: go test -race ./...
```

The `-race` flag instruments your code to detect data races at runtime.
If two goroutines access the same variable without synchronization → test FAILS.

```
WARNING: DATA RACE
Read at 0x00c0000 by goroutine 7:
  main.worker()
Write at 0x00c0000 by goroutine 8:
  main.submit()
```

### C#/Java: No built-in equivalent. You use:
- C#: ThreadSanitizer (external), or manually review `lock` usage
- Java: FindBugs/SpotBugs `@GuardedBy`, or Java Flight Recorder

---

## Deep Dive: Cross-Compilation

Go cross-compiles to ANY platform with just environment variables:

```bash
# Build for Linux ARM (e.g., AWS Graviton, Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o shipit-linux-arm64 ./cmd/shipit

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o shipit.exe ./cmd/shipit

# Build for macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o shipit-darwin-arm64 ./cmd/shipit
```

**No additional toolchain needed.** The Go compiler has every target built-in.

### C# Equivalent:
```bash
dotnet publish -r linux-arm64 --self-contained  # Requires .NET SDK for that RID
```

### Java Equivalent:
```bash
# Requires GraalVM + native-image (complex setup per platform)
native-image -jar app.jar --target=linux-aarch64
```

---

## Deep Dive: Multi-Stage Docker Build

```dockerfile
# Stage 1: Build (large image with Go toolchain ~1GB)
FROM golang:1.22-alpine AS builder
COPY . .
RUN go build -o shipit ./cmd/shipit

# Stage 2: Run (empty image — just our binary)
FROM scratch
COPY --from=builder /app/shipit /shipit
ENTRYPOINT ["/shipit"]
```

### Result: ~10MB image (just the binary + CA certs)

### C# Equivalent:
```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS builder
RUN dotnet publish -c Release -o /app

FROM mcr.microsoft.com/dotnet/aspnet:8.0-alpine  # Still ~100MB (needs runtime)
COPY --from=builder /app .
```

### Java Equivalent:
```dockerfile
FROM maven:3.9 AS builder
RUN mvn package

FROM eclipse-temurin:21-jre-alpine  # Still ~150MB (needs JVM)
COPY --from=builder /app/target/*.jar app.jar
```

### Why Go images are so small:
Go compiles to a **static binary** — no runtime, no VM, no dependencies.
`FROM scratch` = completely empty base image (0 bytes).

---

## Deep Dive: Build Flags

```bash
go build -ldflags="-s -w" -o shipit ./cmd/shipit
```

| Flag | Effect | Size Impact |
|------|--------|-------------|
| `-s` | Strip symbol table | -25% |
| `-w` | Strip DWARF debug info | -10% |
| `CGO_ENABLED=0` | Pure Go (no C deps) | Enables `FROM scratch` |

---

## Deep Dive: Workflow Features

### Concurrency (Cancel Stale Runs):
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```
If you push 3 times quickly, only the latest run continues. Previous runs are cancelled.

### Matrix Strategy (Build for Multiple Platforms):
```yaml
strategy:
  matrix:
    include:
      - goos: linux
        goarch: amd64
      - goos: windows
        goarch: amd64
      - goos: darwin
        goarch: arm64
```

### Conditional Jobs (Docker only on main):
```yaml
if: github.event_name == 'push' && github.ref == 'refs/heads/main'
```

---

## Trigger: On PR Merge

This pipeline triggers on:
- **Pull request** → runs lint + test + build (validates the code)
- **Push to main** (PR merge) → runs everything + pushes Docker image

```yaml
on:
  push:
    branches: [main]    # After PR is merged
  pull_request:
    branches: [main]    # On PR opened/updated
```

---

## Try It

```bash
# The pipeline runs automatically on push/PR
git add .
git commit -m "feat: lesson 15 - CI/CD with GitHub Actions"
git push origin main

# Check: https://github.com/alluri02/go-shipit/actions

# Or test locally what the pipeline does:
go vet ./...                           # Basic lint
go test -race -cover ./...             # Test with race detector
CGO_ENABLED=0 go build -ldflags="-s -w" -o shipit ./cmd/shipit  # Production build
docker build -t shipit .               # Docker build
```

---

## Key Takeaways

1. **`go test -race`** — built-in race detector. Catches concurrency bugs. Unique to Go.
2. **Cross-compilation is trivial** — `GOOS=linux GOARCH=arm64 go build`. Done.
3. **`FROM scratch` Docker images** — ~10MB. No runtime, no VM, no deps.
4. **`-ldflags="-s -w"`** — strip debug info for smaller binaries.
5. **`CGO_ENABLED=0`** — pure Go binary. Works in empty containers.
6. **Lint + Test + Build + Docker** — the standard Go CI pipeline.
7. **Matrix builds** — compile for every platform in parallel.

---

## The Complete Learning Path Is Done! 🎉

You've built a production-grade Go deployment orchestrator with:
- Hexagonal architecture
- HTTP API with middleware
- Concurrent worker processing
- AI integration
- MySQL persistence
- CLI with Cobra
- CI/CD pipeline

All 15 lessons map Go concepts to C#/Java equivalents you already know.
