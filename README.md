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
| 02 | Types, Structs & Methods | Domain models | [Lesson 02](docs/02-types-structs-methods.md) |
| 03 | Interfaces (implicit) | Port definitions | [Lesson 03](docs/03-interfaces-implicit.md) |
| 04 | Error Handling | Domain service errors | [Lesson 04](docs/04-error-handling.md) |
| 05 | Dependency Injection | Wire services manually | [Lesson 05](docs/05-dependency-injection.md) |
| 06 | HTTP Server (`net/http`) | httpservice scaffold | [Lesson 06](docs/06-http-server.md) |
| 07 | Goroutines & Channels | webhookprocessor workers | [Lesson 07](docs/07-goroutines-channels.md) |
| 08 | Testing (table-driven) | Unit tests | [Lesson 08](docs/08-testing.md) |
| 09 | Configuration & Env | Config loading | [Lesson 09](docs/09-configuration.md) |
| 10 | Database (SQL, no ORM) | MySQL adapter + Skeema | [Lesson 10](docs/10-database.md) |
| 11 | Context & Cancellation | Request lifecycle | [Lesson 11](docs/11-context.md) |
| 12 | Middleware & Composition | Auth, logging, metrics | [Lesson 12](docs/12-middleware.md) |
| 13 | AI Integration | Risk scorer via GitHub Models | [Lesson 13](docs/13-ai-integration.md) |
| 14 | CLI with Cobra | Subcommand routing | [Lesson 14](docs/14-cli.md) |
| 15 | CI/CD | GitHub Actions pipeline | [Lesson 15](docs/15-cicd.md) |

---

## Quick Start

```bash
# Prerequisites: Go 1.22+
go build -o shipit.exe ./cmd/shipit

# Run locally (in-memory — no external dependencies needed)
.\shipit.exe serve
.\shipit.exe demo
.\shipit.exe process
```

---

## Running with Real Infrastructure

### Option A: Docker Compose (local, free)

```bash
docker compose up -d
```

This starts MySQL + Azurite (Azure Storage emulator). Then:

```bash
$env:SHIPIT_DATABASE_URL = "root:devpass@tcp(127.0.0.1:3306)/shipit?parseTime=true"
$env:SHIPIT_QUEUE_CONNECTION_STRING = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1"
$env:SHIPIT_ENV = "development"
.\shipit.exe serve
```

| Service | Local URL | Purpose |
|---------|-----------|---------|
| MySQL 8 | `127.0.0.1:3306` | Deployment persistence |
| Azurite Queue | `127.0.0.1:10001` | Message queue (Azure Queue Storage emulator) |
| Azurite Blob | `127.0.0.1:10000` | Blob storage (future: build artifacts) |
| ShipIt API | `127.0.0.1:8080` | HTTP API |

### Option B: Azure Resources (production)

Provision these resources in Azure:

```bash
# 1. Resource Group
az group create --name rg-shipit --location eastus

# 2. Storage Account (for Queue Storage)
az storage account create \
  --name stshipit$(openssl rand -hex 4) \
  --resource-group rg-shipit \
  --sku Standard_LRS

# 3. Azure Container Registry
az acr create \
  --name acrshipit$(openssl rand -hex 4) \
  --resource-group rg-shipit \
  --sku Basic

# 4. Azure Database for MySQL (Flexible Server)
az mysql flexible-server create \
  --name mysql-shipit \
  --resource-group rg-shipit \
  --admin-user shipit \
  --admin-password <YOUR_PASSWORD> \
  --sku-name Standard_B1ms \
  --tier Burstable

# 5. Azure Container Apps Environment (deployment target)
az containerapp env create \
  --name cae-shipit \
  --resource-group rg-shipit \
  --location eastus
```

Then set environment variables:

```bash
$env:SHIPIT_ENV = "production"
$env:SHIPIT_DATABASE_URL = "<mysql-connection-string>"
$env:SHIPIT_QUEUE_CONNECTION_STRING = "<storage-account-connection-string>"
$env:SHIPIT_SLACK_TOKEN = "<slack-bot-token>"
.\shipit.exe serve
```

### What Each Resource Does

| Azure Resource | ShipIt Adapter | Purpose |
|---------------|---------------|---------|
| **Storage Account** (Queue) | `adapters/queue/` | Deploy job queue — webhookreceiver enqueues, processor dequeues |
| **Container Registry** | `adapters/acr/` | Stores built Docker images before deployment |
| **MySQL Flexible Server** | `adapters/mysql/` | Persists deployments, environments, history |
| **Container Apps** | `adapters/deployer/` | Target for deploying container images |
| **GitHub Models** | `adapters/ai/` | GPT-4o for deploy risk scoring (uses GitHub token) |

### Remaining Adapters to Implement

| Adapter | Status | What's Needed |
|---------|--------|---------------|
| `inmemory/` | ✅ Done | Works now — no infra needed |
| `mysql/` | ✅ Done | Just needs a running MySQL |
| `ai/` | ✅ Done | Just needs `GITHUB_TOKEN` |
| `queue/azure` | 🔲 TODO | Azure Storage SDK (`azqueue`) |
| `acr/` | 🔲 TODO | Docker build + ACR push |
| `deployer/containerapp` | 🔲 TODO | Azure Container Apps SDK |
| `slack/` | 🔲 TODO | Slack Bot API |

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
