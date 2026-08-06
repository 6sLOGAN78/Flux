# CODING_STANDARDS.md — Code Quality & Monorepo Engineering Conventions

## 1. Monorepo Layout Guidelines
- Root directory (`/`) contains workspace configurations (`package.json`, `turbo.json`, `docker-compose.yml`) and root documentation (`AGENTS.md`, `README.md`, `PRODUCT.md`).
- Core services belong in `apps/` (`apps/backend`, `apps/frontend`).
- Shared modules belong in `packages/` (`packages/zod`, `packages/openapi`, `packages/emails`).

## 2. Go Backend Guidelines (`apps/backend`)
- Use **Echo v4** (`labstack/echo/v4`) for HTTP routing, handlers, and middlewares.
- Database access MUST use `pgx/v5` (`jackc/pgx/v5`) connection pooling with raw SQL queries.
- Migrations MUST be written as raw SQL files under `apps/backend/internal/database/migrations/` in **Tern** (`jackc/tern/v2`) format.
- Background asynchronous processing MUST use **Asynq** (`hibiken/asynq`) on **Redis 8**.
- Use explicit type declarations and structured error wrapping (`fmt.Errorf("failed to fetch link: %w", err)`).
- Validate HTTP request structs using `go-playground/validator/v10`.
- Execute `Taskfile.yml` commands (`task test`, `task lint`) before committing code.

## 3. TypeScript & React Guidelines (`apps/frontend`, `packages/*`)
- Use **Bun** (`bun@1.2.13`) as the JS/TS package manager.
- Entity schemas MUST be defined in `packages/zod` and exported as inferred TypeScript types.
- API contracts MUST be defined in `packages/openapi` using `@ts-rest/core`.
- Frontend API interactions MUST use the typed `@ts-rest` client coupled with `@tanstack/react-query`.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Previous** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md) |
| **Next** | [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Children** | None |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
