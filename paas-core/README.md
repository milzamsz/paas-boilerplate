# paas-core

PaaS Core — Go API backend + Next.js Dashboard + shared packages.

## Structure

```
paas-core/
├── apps/
│   ├── api/               # Go API (Chi + pgx)
│   │   ├── cmd/server/    # Entry point
│   │   ├── internal/      # Private packages
│   │   │   ├── config/    # Environment config
│   │   │   ├── database/  # PostgreSQL pool
│   │   │   ├── handler/   # HTTP handlers
│   │   │   ├── middleware/ # JWT, RBAC, org resolver
│   │   │   ├── model/     # Domain models
│   │   │   └── response/  # JSON response helpers
│   │   ├── migrations/    # SQL migrations
│   │   └── docs/          # OpenAPI spec (generated)
│   └── web/               # Next.js 15 Dashboard
├── packages/
│   └── api-client/        # TypeScript API client (Orval)
├── deploy/
│   └── docker/            # Dockerfiles + Compose
├── scripts/               # Dev scripts
└── .github/workflows/     # CI/CD
```

## Quick Start

```bash
# 1. Clone and configure
cp .env.example .env

# 2. Start everything (Docker)
make dev

# 3. Or run individually
make api-dev    # Go API on :8080
make web-dev    # Next.js on :3000
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/auth/register` | Register user |
| `POST` | `/api/v1/auth/login` | Login |
| `POST` | `/api/v1/auth/refresh` | Refresh JWT |
| `GET` | `/api/v1/user/me` | Current user profile |
| `PUT` | `/api/v1/user/me` | Update profile |
| `GET` | `/api/v1/orgs` | List user organizations |
| `POST` | `/api/v1/orgs` | Create organization |
| `GET` | `/api/v1/orgs/{orgId}` | Get organization |
| `PUT` | `/api/v1/orgs/{orgId}` | Update organization |
| `DELETE` | `/api/v1/orgs/{orgId}` | Delete organization |
| `GET` | `/api/v1/orgs/{orgId}/members` | List members |
| `POST` | `/api/v1/orgs/{orgId}/members` | Invite member |
| `PUT` | `/api/v1/orgs/{orgId}/members/{memberId}` | Update role |
| `DELETE` | `/api/v1/orgs/{orgId}/members/{memberId}` | Remove member |
| `GET` | `/api/v1/orgs/{orgId}/projects` | List projects |
| `POST` | `/api/v1/orgs/{orgId}/projects` | Create project |
| `GET` | `/api/v1/orgs/{orgId}/projects/{projectId}` | Get project |
| `PUT` | `/api/v1/orgs/{orgId}/projects/{projectId}` | Update project |
| `DELETE` | `/api/v1/orgs/{orgId}/projects/{projectId}` | Delete project |
| `POST` | `/api/v1/orgs/{orgId}/projects/{projectId}/deploy` | Create deployment |
| `GET` | `/api/v1/orgs/{orgId}/projects/{projectId}/deployments` | List deployments |
| `GET` | `/api/v1/orgs/{orgId}/projects/{projectId}/deployments/{deployId}` | Get deployment |
| `GET` | `/api/v1/orgs/{orgId}/projects/{projectId}/env` | List env vars |
| `POST` | `/api/v1/orgs/{orgId}/projects/{projectId}/env` | Upsert env var |
| `GET` | `/api/v1/orgs/{orgId}/billing` | Billing overview |
| `GET` | `/api/v1/orgs/{orgId}/billing/plans` | List plans |
| `POST` | `/api/v1/orgs/{orgId}/billing/subscribe` | Subscribe |
| `POST` | `/api/v1/orgs/{orgId}/billing/cancel` | Cancel subscription |
| `GET` | `/api/v1/orgs/{orgId}/billing/invoices` | List invoices |
| `GET` | `/api/v1/orgs/{orgId}/billing/usage` | Usage metrics |
| `GET` | `/api/v1/orgs/{orgId}/audit-logs` | Audit logs |
| `POST` | `/api/v1/webhooks/xendit` | Xendit webhook |
| `GET` | `/healthz` | Health check |
| `GET` | `/readyz` | Readiness check |

## Tech Stack

- **API**: Go 1.22+, Chi v5, pgx v5
- **Dashboard**: Next.js 15, TypeScript
- **Database**: PostgreSQL 16
- **Auth**: JWT (Bearer + HttpOnly cookie)
- **Billing**: Xendit
- **CI**: GitHub Actions
