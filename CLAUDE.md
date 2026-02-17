# CLAUDE.md — paas-boilerplate

## Project Overview

Multi-tenant PaaS (Platform-as-a-Service) boilerplate workspace containing separate repos:

- **paas-core/** — Go API + Next.js Dashboard (fully implemented)
- **paas-site/** — Marketing website (not yet created)
- **paas-docs/** — Documentation site (not yet created)

## Tech Stack

| Layer       | Technology                                    |
| ----------- | --------------------------------------------- |
| API         | Go 1.24, Gin, GORM, PostgreSQL 16             |
| Dashboard   | Next.js 15, React 19, shadcn/ui, Tailwind CSS |
| Auth        | JWT (access + refresh tokens)                 |
| Billing     | Xendit payment gateway                        |
| Email       | Resend                                        |
| Storage     | S3-compatible (AWS/MinIO)                     |
| API Client  | Orval (TypeScript, generated from OpenAPI)    |
| Containers  | Docker multi-stage builds, Docker Compose     |
| CI/CD       | GitHub Actions                                |

## Project Structure

```
paas-boilerplate/
├── STACK.config.json           # Workspace-level config (product name, domains, etc.)
├── paas-core/
│   ├── apps/
│   │   ├── api/                # Go REST API
│   │   │   ├── cmd/server/     # main.go entry point
│   │   │   ├── internal/       # auth, billing, config, database, email, errors,
│   │   │   │                   # featuregate, middleware, model, oauth, org,
│   │   │   │                   # project, storage, user
│   │   │   ├── migrations/     # SQL migration files
│   │   │   ├── docs/           # Generated OpenAPI spec
│   │   │   └── go.mod          # Go module: paas-core/apps/api
│   │   └── web/                # Next.js 15 dashboard
│   ├── packages/
│   │   └── api-client/         # Orval-generated TypeScript API client
│   ├── deploy/docker/          # docker-compose.yml, Dockerfiles
│   ├── Makefile                # Dev commands
│   └── .env.example            # Environment template
```

## Common Commands

```bash
# From paas-core/
make dev              # Docker Compose full stack (API + Web + DB)
make dev-down         # Stop all containers
make api-dev          # Run Go API locally (port 8080)
make web-dev          # Run Next.js locally (port 3000)
make api-test         # Run Go tests
make web-test         # Run Next.js tests
make api-lint         # go vet
make web-lint         # pnpm lint
make openapi          # Generate OpenAPI spec (swag)
make gen-client       # Generate TypeScript client (Orval)
make build            # Build Docker images (api + web)
```

## Architecture

- **Multi-tenancy:** Single shared database, scoped by `org_id` on all domain tables
- **API routing:** Path-based — `/api/v1/orgs/{orgId}/...`
- **Auth flow:** JWT access token (15m) + refresh token (7d), HttpOnly cookies
- **RBAC roles:** super_admin, admin, owner, developer, viewer
- **Middleware chain:** Recovery → RequestID → Logger → ErrorHandler → SecurityHeaders → CORS → CSRF → JWTAuth → OrgResolver

## API Route Groups

- `POST /api/v1/auth/{register,login,refresh}` — Public, rate-limited
- `GET/PUT /api/v1/users/me` — Authenticated user profile
- `POST/GET /api/v1/orgs` — Org CRUD
- `/api/v1/orgs/:orgId/members` — Member management
- `/api/v1/orgs/:orgId/projects` — Project CRUD + deployments + env vars
- `/api/v1/orgs/:orgId/billing` — Subscription, invoices, usage
- `POST /api/v1/webhooks/xendit` — Billing webhooks (signature-verified)
- `GET /healthz`, `GET /readyz` — Health checks

## Key Configuration

- **STACK.config.json** — Product name, domains, deploy target, tenancy mode
- **.env.example** — Database URL, JWT secret, CORS origins, Xendit keys, S3 config
- **apps/api/configs/** — YAML config files loaded by Viper

## Database

- PostgreSQL 16, GORM auto-migrate on startup
- Models: User, Role, UserRole, RefreshToken, EmailVerificationToken, PasswordResetToken, OAuthAccount, FileUpload, Org, Membership, Project, Deployment, BillingPlan, Subscription, Invoice, AuditLog
- Default connection: `postgres://paas:paas@localhost:5432/paas`

## Code Conventions

- Go: standard library `log/slog` for logging, `internal/` for private packages
- Go: repository → service → handler pattern per domain
- Frontend: Next.js App Router, server components where possible
- Naming: kebab-case files, PascalCase types, camelCase functions (Go exported = PascalCase)
