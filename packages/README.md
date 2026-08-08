# Workspace Shared Packages (`packages/`)

This directory contains shared packages published across the monorepo:

---

## 📦 Workspace Packages

- **`@flux/openapi`**: Bundles the static OpenAPI contract JSON (`openapi.json`) for frontend client generation and API consumers.
- **`@flux/zod`**: Provides type-safe **Zod schemas** (`ZLink`, `ZCategory`, `ZCampaign`, `ZCustomDomain`, `ZAttributionResult`, etc.) and inferred TypeScript types shared between frontend and backend contracts.

---

## 🚀 Usage

In any monorepo package or app:

```typescript
import { ZLink, Link, ZAttributionResult } from "@flux/zod";
```
