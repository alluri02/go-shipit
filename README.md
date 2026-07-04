# go-shipit

> **Learn production-grade Go by building a real deployment orchestrator — with Java/C# concept mapping**

This isn't a toy project. It's a real system with hexagonal architecture, MySQL, message queues, AI integration, and CI/CD — the same patterns used at GitHub, Google, and Uber in their Go services.

---

## Who Is This For?

Engineers with **Java or C# experience** who want to learn Go by building something real. Every commit teaches one Go concept, mapped back to what you already know.

---

## What We're Building: ShipIt

A deployment pipeline orchestrator that:
- Receives GitHub webhooks and Slack commands
- Queues deployment jobs via Azure Queue Storage
- Builds and pushes container images to ACR
- Deploys to Azure Container Apps
- Scores deploy risk using AI (GitHub Models)

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Triggers: Slack ChatOps │ GitHub Webhooks │ REST Clients   │
└─────────────────┬───────────────────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────────────────┐
│  Transport Layer (single binary, 4 subcommands)             │
│  httpservice │ chatopsservice │ webhookreceiver │ processor │
└─────────────────┬───────────────────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────────────────┐
│  Domain Layer (pure business logic — zero external deps)    │
│  DeployService │ BuildService │ RiskService │ ...           │
└─────────────────┬───────────────────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────────────────┐
│  Adapters (MySQL │ Azure Queue │ ACR │ GitHub │ Slack │ AI) │
└─────────────────────────────────────────────────────────────┘
```

---

## Learning Path (each commit = one lesson)

| # | Go Concept | ShipIt Progress | Doc |
|---|-----------|-----------------|-----|
| 01 | Modules & Packages | Project init, folder structure | [Lesson 01](docs/01-modules-and-packages.md) |
| 02 | Types, Structs & Methods | Domain models | Coming next |
| 03 | Interfaces (implicit) | Port definitions | — |
| 04 | Error Handling | Domain service errors | — |
| 05 | Dependency Injection | Wire services manually | — |
| 06 | HTTP Server (`net/http`) | httpservice scaffold | — |
| 07 | Goroutines & Channels | webhookprocessor workers | — |
| 08 | Testing (table-driven) | Unit tests | — |
| 09 | Configuration & Env | Config loading | — |
| 10 | Database (SQL, no ORM) | MySQL adapter + Skeema | — |
| 11 | Context & Cancellation | Request lifecycle | — |
| 12 | Middleware & Composition | Auth, logging, metrics | — |
| 13 | AI Integration | Risk scorer via GitHub Models | — |
| 14 | CLI with Cobra | Subcommand routing | — |
| 15 | CI/CD | GitHub Actions pipeline | — |

---

## Quick Start

```bash
# Prerequisites: Go 1.21+
go run ./cmd/shipit httpservice
```

---

## Project Structure

```
go-shipit/
├── cmd/shipit/           ← Entry point (package main)
├── internal/
│   ├── domain/           ← Business logic (no deps)
│   ├── ports/            ← Interfaces (future)
│   └── adapters/         ← External integrations (future)
├── docs/                 ← One doc per lesson
├── go.mod                ← Module definition
└── README.md
```

---

## Go ↔ C# ↔ Java Quick Reference

| Concept | Go | C# | Java |
|---------|-----|-----|------|
| Project file | `go.mod` | `.csproj` | `pom.xml` |
| Package | `package domain` (folder) | `namespace X` | `package x` |
| Entry point | `func main()` | `static void Main()` | `public static void main()` |
| Public | `Uppercase` | `public` keyword | `public` keyword |
| Private | `lowercase` | `private` keyword | `private` keyword |
| Unused import | Compile error | Warning | Warning |
| Null | `nil` | `null` | `null` |
| Inheritance | None (composition only) | `class : Base` | `class extends Base` |
| Interface impl | Implicit (no keyword) | `class : IFoo` | `class implements Foo` |

---

## License

MIT
