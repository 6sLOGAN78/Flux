# Development Guide

## Prerequisites
- Go 1.25+
- Node 20+ / Bun 1.1+
- Docker & Docker Compose

## Running Locally
1. Start infrastructure:
   ```bash
   docker compose up -d postgres redis
   ```
2. Start backend:
   ```bash
   cd apps/backend
   go run ./cmd/api/main.go
   ```
3. Start frontend (in another terminal):
   ```bash
   bun run dev
   ```

## Note on Architecture
Most advanced features (campaigns, billing, domains) are currently **UI-only mocks**. The Go backend contains logic in `internal/modules` but they are not wired to the database or the HTTP router.
