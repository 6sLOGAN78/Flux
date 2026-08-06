# Flux — Multi-Tier Link Management & Enterprise Attribution Platform

Flux is a high-performance, production-grade URL shortening, dynamic link routing, real-time analytics, SaaS monetization, and multi-touch attribution platform built in Go.

## 📁 Repository Structure

```text
Flux/
├── README.md              ← Project overview & sitemap
├── AGENTS.md              ← Master AI Agent Operating Protocol
├── PRODUCT.md             ← Product vision & feature matrix
├── ROADMAP.md             ← Tier evolution roadmap
├── CHANGELOG.md           ← Version changelog
├── CONTRIBUTING.md        ← Contribution guidelines
│
├── ai/                    ← AI Agent Adapters & Engineering Standards
│   ├── CLAUDE.md          ← Claude Code CLI adapter rules
│   ├── GEMINI.md          ← Gemini AI Agent adapter rules
│   ├── CURSOR.md          ← Cursor IDE agent adapter rules
│   ├── CODEX.md           ← OpenAI Codex CLI adapter rules
│   ├── ANTIGRAVITY.md     ← Google Antigravity agent adapter rules
│   ├── CODING_STANDARDS.md← Code quality & engineering conventions
│   ├── CONVENTIONS.md     ← General agent conventions & SRP rules
│   ├── SECURITY.md        ← Security architecture & auth policies
│   ├── PERFORMANCE.md     ← SLAs, latency budgets & caching strategy
│   ├── TESTING.md         ← Testing strategy & TDD mandate
│   ├── DEPLOYMENT.md      ← Packaging & production deployment
│   └── STYLE_GUIDE.md     ← Code formatting & style guidelines
│
├── docs/                  ← Technical subsystem specifications
├── api/                   ← OpenAPI & AsyncAPI schema contracts
├── database/              ← PostgreSQL master schema contracts
├── tasks/                 ← Executable task units
├── templates/             ← Architectural & task document templates
├── adr/                   ← Architecture Decision Records
└── src/                   ← Core Go backend implementation
    ├── cmd/               ← Application entrypoints (API server, worker)
    └── internal/          ← Internal application modules & configuration
```

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 15+

### Build & Run
```bash
# Run unit tests
go test -v ./...

# Build API server
go build -o bin/flux-api ./src/cmd/api

# Run API server
./bin/flux-api
```

## 📜 AI Agent Protocol
All AI coding agents operating in this repository MUST strictly follow the operating protocol defined in [`AGENTS.md`](file:///home/logan78/Desktop/flux/AGENTS.md) and reference documents in [`ai/`](file:///home/logan78/Desktop/flux/ai).
