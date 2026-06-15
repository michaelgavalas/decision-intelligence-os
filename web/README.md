# @dios/web

The web client for **Decision Intelligence OS** - a decision-quality platform for
structured forecasting, assumption tracking, evidence collection, and outcome
analysis.

## Stack

- **TypeScript** + **React** (Vite)
- **Apollo Client** for the GraphQL data layer
- **React Router** for routing
- **Tailwind CSS v4** with a tokenized, class-based light/dark theme
- **Radix UI** primitives, **lucide-react** icons, **Recharts** for analytics
- **Vitest** + **React Testing Library** + **MSW** for tests
- **graphql-codegen** for a fully typed GraphQL API surface

The UI follows a feature-based architecture (`src/features/*`) with shared,
shadcn-style primitives in `src/components/ui`.

## Prerequisites

- Node 24+
- pnpm 11+
- The Go GraphQL backend running locally on `http://localhost:8080`
  (see [`../backend`](../backend))

This package is a member of the repository's pnpm workspace, so install
dependencies from the repository root.

## Getting started

```bash
# From the repository root
pnpm install

# Generate typed GraphQL hooks from the backend schema
pnpm --filter @dios/web codegen

# Start the dev server (proxies /graphql to the backend)
pnpm --filter @dios/web dev
```

Copy `.env.example` to `.env` if you need to override defaults. In development
`VITE_GRAPHQL_URL` defaults to `/graphql`, which Vite proxies to the backend so
that the httpOnly refresh cookie and CSRF token remain same-origin.

## Scripts

| Script      | Description                                     |
| ----------- | ----------------------------------------------- |
| `dev`       | Start the Vite dev server                       |
| `build`     | Type-check the project and build for production |
| `preview`   | Preview the production build locally            |
| `lint`      | Run ESLint                                      |
| `typecheck` | Type-check without emitting                     |
| `test`      | Run the Vitest suite                            |
| `codegen`   | Regenerate typed GraphQL operations             |

Run any script through the workspace, e.g. `pnpm --filter @dios/web test`.

## Project layout

```text
src/
├── components/   # Shared UI primitives (ui/) and chart wrappers (charts/)
├── features/     # Feature modules (auth, decisions, analytics, teams, ...)
├── graphql/      # GraphQL operations and generated types
├── hooks/        # Reusable hooks
├── layouts/      # App shell and auth layouts
├── lib/          # Apollo client, auth token store, helpers
├── pages/        # Route-level placeholder pages
├── styles/       # Global styles and design tokens
└── test/         # Test setup and MSW handlers
```

## Authentication model

Access tokens are held in memory only (never `localStorage`) and attached as a
bearer header by Apollo. The refresh token lives in an httpOnly cookie; Apollo
transparently refreshes an expired access token once per failed request using
the CSRF-protected `refreshToken` mutation, then replays the original operation.
