# PLAN-011: Frontend Scaffold + Monorepo Setup

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-16 |
| **Depends on** | — |

## Goal

Add a `frontend/` directory to the monorepo and bootstrap a modern React SPA that will serve as the
admin panel. The backend Go code stays in its current locations — only the directory layout and
tooling are extended, not restructured.

## Monorepo Layout After This Plan

```
event-sourcing-learning/
  wallet-service/          ← unchanged
  kyc-service/             ← unchanged
  contracts/               ← unchanged
  frontend/                ← NEW
    src/
      api/                 ← generated API client (openapi-fetch)
      components/          ← shared UI primitives
      pages/               ← route-level page components
      hooks/               ← TanStack Query hooks
      lib/                 ← utils, constants
    public/
    index.html
    package.json
    vite.config.ts
    tsconfig.json
    tailwind.config.ts
    components.json        ← shadcn/ui config
    .eslintrc.cjs
    .prettierrc
  docker-compose.yml       ← add frontend dev service
  go.work
```

## Tech Stack

| Concern | Choice | Reason |
|---------|--------|--------|
| Bundler | Vite 5 | Fast HMR, native ESM, first-class TS |
| Framework | React 18 | Ecosystem, shadcn/ui support |
| Language | TypeScript 5 (strict) | Type safety end-to-end |
| Routing | TanStack Router v1 | Type-safe routes, file-based, SSR-ready |
| Data fetching | TanStack Query v5 | Caching, background refetch, devtools |
| UI components | shadcn/ui + Radix | Accessible, unstyled primitives, copy-paste |
| Styling | Tailwind CSS v3 | Utility-first, consistent design tokens |
| API client | openapi-fetch | Type-safe fetch generated from OpenAPI spec |
| Package manager | pnpm | Fast, disk-efficient, strictness |
| Linting | ESLint + typescript-eslint | Catch TS anti-patterns |
| Formatting | Prettier | Zero-config formatting |

## Vite Dev Proxy

`vite.config.ts` proxies `/api` → `http://localhost:8080` so the SPA
talks to the local wallet-service without CORS issues in dev:

```ts
server: {
  proxy: {
    '/api': { target: 'http://localhost:8080', changeOrigin: true },
  },
},
```

## Initial Route Structure

```
/login               ← public, no auth required
/                    ← redirect → /dashboard
/dashboard           ← protected
/accounts            ← protected
/accounts/:id        ← protected
/projector           ← protected
```

TanStack Router uses file-based routing under `src/pages/`:
```
src/pages/
  login.tsx
  _authenticated/         ← layout with auth guard
    dashboard.tsx
    accounts/
      index.tsx
      $accountId.tsx
    projector.tsx
```

## Acceptance Criteria

- [ ] `pnpm install` completes without errors
- [ ] `pnpm dev` starts Vite dev server on port 5173
- [ ] `pnpm build` produces a dist/ with no TS errors
- [ ] `pnpm lint` passes with zero warnings
- [ ] Proxy `/api/*` → `localhost:8080` works in dev mode
- [ ] TanStack Router devtools visible in dev mode
- [ ] shadcn/ui `Button`, `Card`, `Table` components importable
- [ ] Basic login page renders (static, no API calls yet)
- [ ] `_authenticated` layout renders a nav sidebar + outlet

## Tasks

- [ ] `pnpm create vite frontend -- --template react-ts`
- [ ] Add TanStack Router, TanStack Query, openapi-fetch
- [ ] Configure Tailwind CSS + shadcn/ui init
- [ ] Configure Vite proxy for `/api`
- [ ] Set up TanStack Router with file-based routing
- [ ] Create `_authenticated` layout (sidebar nav + outlet + auth guard stub)
- [ ] Create static login page (form, no API yet)
- [ ] Add ESLint + prettier config
- [ ] Add `pnpm dev` / `pnpm build` / `pnpm lint` scripts
- [ ] Update docs
