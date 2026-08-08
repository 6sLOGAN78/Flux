# Flux Frontend Web Application (`apps/frontend`)

The **Flux Frontend** is a modern dashboard web application built with **React 19**, **TypeScript**, **Vite**, and **Tailwind CSS**.

---

## 🛠 Tech Stack

- **Framework**: React 19 + Vite 6
- **Language**: TypeScript 5.7
- **Styling**: Tailwind CSS v4 + `clsx` + `tailwind-merge`
- **Data Fetching**: `@tanstack/react-query` v5
- **Authentication**: `@clerk/clerk-react` v5
- **Icons**: `lucide-react`
- **Workspace Dependencies**: `@flux/openapi` and `@flux/zod`

---

## 📂 Project Structure

```text
apps/frontend/
├── src/
│   ├── components/         ← Reusable UI components (Buttons, Inputs, Cards, Modals)
│   ├── pages/              ← Dashboard pages (Overview, Links, Analytics, Domains, Settings)
│   ├── hooks/              ← Custom React hooks & TanStack Query integrations
│   ├── lib/                ← API client wrappers & helper utilities
│   └── App.tsx             ← Application entrypoint & router setup
├── package.json            ← Node dependencies & npm scripts
└── vite.config.ts          ← Vite build configuration
```

---

## 🚀 Getting Started

### 1. Install Dependencies
From the workspace root or frontend directory:
```bash
bun install
```

### 2. Run Development Server
```bash
bun run dev
```
The development server will start on `http://localhost:5173`.

### 3. Build for Production
```bash
bun run build
```
Outputs static assets into `dist/`.
