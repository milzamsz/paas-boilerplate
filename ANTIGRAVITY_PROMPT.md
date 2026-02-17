Referensi stack yang kamu pilih (biar Antigravity bisa “grounded” saat nge-clone/inspeksi): **GRAB** ([GitHub][1]), **AstroWind** ([GitHub][2]), **Fumadocs UI template** ([GitHub][3]), dan **fumadocs-openapi** ([npm][4]).

Di bawah ini **prompt template** (copy–paste) yang bisa kamu taruh di root repo (mis. `ANTIGRAVITY_PROMPT.md`) dan langsung kamu lempar ke Antigravity.

```md
# Antigravity Master Prompt — Full Stack Template (PaaS Multi-Tenant)
Kamu adalah Senior Tech Lead + DevOps + Fullstack Engineer. Tugasmu: buat **workspace template** untuk produk **PaaS multi-tenant** berbasis Docker, dengan 3 repo terpisah:
1) `paas-core` (private): Monorepo **GRAB (Go API)** + **Next.js Dashboard (Kiranism shadcn)**  
2) `paas-site` (public): Marketing website pakai **AstroWind (Astro + Tailwind)**  
3) `paas-docs` (public): Documentation pakai **Fumadocs** (basis dari `fumadocs-ui-template`) + auto-generate API reference dari OpenAPI.

> Constraint: Docker-first, siap deploy di PaaS (Dokploy/Coolify sejenis).  
> Prinsip: **split repo** untuk `site` dan `docs`, core aplikasi tetap terpisah dan aman.

---

## 0) Parameter (gunakan default kalau tidak tersedia)
Buat file `STACK.config.json` di workspace root dengan nilai default berikut (boleh diubah user belakangan):
- PRODUCT_NAME: "MyPaaS"
- ORG_NAME: "myorg"
- ROOT_DOMAIN: "example.com"
- SUBDOMAIN_APP: "app"
- SUBDOMAIN_API: "api"
- SUBDOMAIN_DOCS: "docs"
- SUBDOMAIN_SITE: "" (root domain)
- DEPLOY_TARGET: "docker" (opsional: "dokploy" | "coolify" | "k8s-later")
- TENANCY_MODE: "path" (opsional future: "subdomain")
- AUTH_MODE: "jwt" (opsional future: "oidc")
- DB: "postgres"
- REDIS_ENABLED: false (MVP)
- OPENAPI_SYNC: "release-asset" (opsional: "s3" | "fetch-url")

[ASSUMPTION] Workspace ini bukan 1 git repo tunggal, tapi folder berisi 3 repo. Jangan bikin nested `.git` di root workspace.

---

## 1) OUTPUT WAJIB (folder + dokumen)
Buat struktur workspace:
```

workspace/
paas-core/
paas-site/
paas-docs/
STACK.config.json
README.md               # how to run all
RELEASE_FLOW.md         # cross-repo release & openapi sync
DOMAIN_ROUTING.md       # domain plan (site/app/api/docs)

```

### Acceptance Criteria global
- Masing-masing repo bisa `docker build` dan `docker run` sendiri.
- `paas-core` bisa `docker compose up` untuk local dev (db + api + web).
- `paas-site` build static dan serve via nginx.
- `paas-docs` build Next.js docs site dan punya pipeline generate API docs dari openapi.

---

## 2) Repo A — `paas-core` (Private): GRAB + Next Dashboard Monorepo
### 2.1 Bootstrap & Structure
Buat repo `paas-core` dengan struktur:
```

paas-core/
apps/
api/                  # GRAB source (Go)
web/                  # Next.js dashboard (Kiranism starter)
packages/
api-client/           # TS client generated from OpenAPI (orval)
scripts/
gen-client.sh
deploy/
docker/
docker-compose.yml
Makefile
.env.example
README.md
.github/workflows/ci.yml

```

Rules:
- Monorepo source, tapi **deploy 2 image**: `api` dan `web`.
- Jangan campur runtime. `apps/api` pure Go module. `apps/web` pure Next.
- FE konsumsi API via `packages/api-client` yang di-generate dari OpenAPI.

### 2.2 Multi-Tenant MVP (wajib)
Implementasikan tenancy: **single DB shared schema** + `org_id` di semua tabel domain.

Wajib ada domain model minimum:
- users
- orgs
- org_memberships (role per org)
- org_invites
- projects/apps
- deployments (placeholder)
- env_vars/secrets (placeholder, minimal)
- audit_logs

Tenancy Resolver:
- default route scoping: `/api/v1/orgs/{orgId}/...`
- user bisa switch org di FE (activeOrgId).

RBAC roles:
- owner, admin, developer, viewer  
Enforce RBAC di backend (middleware/policy layer) dan FE (route gating).

### 2.3 API Spec & OpenAPI
- Pastikan API expose OpenAPI JSON (mis. `/api/v1/openapi.json`) dan/atau file build output `apps/api/docs/openapi.json`.
- Standarisasi error format: `{ code, message, details?, request_id }`
- Logging JSON + request-id middleware.

### 2.4 FE Dashboard (Kiranism)
- Sidebar tree minimal:
  - Overview
  - Projects/Apps
  - Deployments
  - Logs (placeholder)
  - Domains (placeholder)
  - Env/Secrets
  - Members & Roles
  - Audit Logs
  - Settings

Auth FE:
- JWT via HttpOnly cookie (preferred) atau token store aman.
- Route protection via Next middleware.

### 2.5 OpenAPI → TS Client (Orval)
- `packages/api-client` berisi config Orval dan output codegen.
- `scripts/gen-client.sh` melakukan:
  1) generate/export openapi dari backend
  2) run orval untuk update client
- Tambahkan `make gen-client` dan `make openapi`.

### 2.6 Docker & Dev UX
- `docker-compose.yml` di root `paas-core` untuk local dev:
  - postgres
  - api
  - web
  - redis optional off by default
- `.env.example` lengkap: DB URL, JWT secrets, CORS, base URLs.
- Makefile target minimal:
  - `make dev`
  - `make api-test`
  - `make web-test`
  - `make openapi`
  - `make gen-client`
  - `make lint`

### 2.7 CI
- GitHub Actions `ci.yml`:
  - path filter: perubahan `apps/api/**` → job api; `apps/web/**`/`packages/**` → job web
  - job validate `gen-client` (optional tapi recommended)
  - caching untuk Go + pnpm
- Output minimal: build + test. Registry push optional (buat placeholder).

---

## 3) Repo B — `paas-site` (Public): AstroWind (Astro + Tailwind)
### 3.1 Bootstrap
Buat repo `paas-site` dari template AstroWind (clone/copy).
- Update branding placeholder (logo/text) pakai `PRODUCT_NAME`.
- Set CTA links:
  - Primary: `https://{SUBDOMAIN_APP}.{ROOT_DOMAIN}`
  - Docs: `https://{SUBDOMAIN_DOCS}.{ROOT_DOMAIN}`
  - Status (optional): `https://status.{ROOT_DOMAIN}`

### 3.2 Content structure (minimum pages)
- Home
- Features
- Pricing
- Security (penting buat PaaS)
- Changelog (optional)
- Contact

[ASSUMPTION] Ini marketing-only, no auth, static-first.

### 3.3 Docker (Static Build)
- Gunakan `astro build` → serve hasil `dist/` via Nginx.
- Buat:
  - `Dockerfile` (multi-stage build) + nginx runtime
  - `.env.example` untuk URL app/docs/api (kalau butuh)
- Tambahkan `README.md` yang jelas cara local dev & docker run.

### 3.4 CI
- Workflow build/test basic.
- Deploy pipeline placeholder (optional).

---

## 4) Repo C — `paas-docs` (Public): Fumadocs + OpenAPI Autogen
### 4.1 Bootstrap
Buat repo `paas-docs` dari `fumadocs-ui-template` sebagai basis.
- Struktur docs minimal:
  - Getting Started
  - Concepts (Tenancy, RBAC, Projects, Deployments)
  - Guides (Deploy app, Configure domain, Secrets)
  - API Reference (generated)

### 4.2 OpenAPI integration (mandatory)
Gunakan `fumadocs-openapi` untuk generate MDX API docs dari schema OpenAPI.

Implementasikan strategi `OPENAPI_SYNC = "release-asset"`:
- `paas-core` saat release/tag:
  - generate `openapi.json`
  - publish sebagai GitHub Release Asset bernama `openapi.json`
- `paas-docs` build pipeline:
  - download `openapi.json` dari release terbaru (atau dari tag tertentu)
  - run generator (fumadocs-openapi) untuk output MDX ke folder API reference
  - build Next docs site

Tambahkan:
- `scripts/fetch-openapi.ts` atau `.sh` untuk download openapi sesuai config:
  - env `OPENAPI_SOURCE` = github-release | url | local
  - env `OPENAPI_VERSION` = latest | vX.Y.Z
- Dokumentasikan di `README.md`.

[DECISION NEEDED] Kalau user tidak mau GitHub Releases, sediakan opsi B:
- publish openapi ke S3/MinIO public bucket dan docs fetch dari URL.

### 4.3 Docker
- Buat Dockerfile Next.js standard (build → run).
- `.env.example` (OPENAPI source, base URL api playground).

### 4.4 CI
- Workflow:
  - lint/test
  - fetch openapi + generate MDX
  - build docs

---

## 5) Cross-Repo Documents (wajib dibuat di workspace root)
### 5.1 `DOMAIN_ROUTING.md`
Tuliskan routing plan:
- `https://{ROOT_DOMAIN}` → paas-site
- `https://{SUBDOMAIN_APP}.{ROOT_DOMAIN}` → paas-core/apps/web
- `https://{SUBDOMAIN_API}.{ROOT_DOMAIN}` → paas-core/apps/api
- `https://{SUBDOMAIN_DOCS}.{ROOT_DOMAIN}` → paas-docs

CORS:
- Allow origin `app` dan `site` jika perlu (site biasanya tidak call api langsung).
- Cookie domain policy untuk auth (kalau app & api beda subdomain).

### 5.2 `RELEASE_FLOW.md`
Flow release versi:
- `paas-core` tag `vX.Y.Z` → publish docker images + openapi asset
- `paas-docs` update untuk version tertentu:
  - pilih `latest` atau follow tag
- `paas-site` independent (konten marketing).

---

## 6) DoD Checklist (wajib di akhir)
Tambahkan checklist final di masing-masing repo README:
- Local dev works
- Docker build works
- Basic lint/test works
- Multi-tenant flows (core): create org, invite member, switch org, RBAC enforced
- Docs: API reference generated from openapi
- Site: CTA links correct

---

## 7) Eksekusi
Sekarang lakukan:
1) Buat folder `workspace/` sesuai struktur.
2) Generate 3 repo sesuai spesifikasi.
3) Pastikan semua README + env example ada.
4) Pastikan semua diagram/arsitektur ditulis jelas (boleh Mermaid di README).
5) Jangan tinggalkan placeholder penting tanpa tanda `[TODO]` yang jelas.

Keluarkan output final berupa:
- daftar file yang dibuat
- cara run masing-masing repo (perintah singkat)
- catatan asumsi & decision needed yang belum diputuskan user.
```


[1]: https://github.com/vahiiiid/go-rest-api-boilerplate?utm_source=chatgpt.com "vahiiiid/go-rest-api-boilerplate: Production-ready AI-friendly ... - GitHub"
[2]: https://github.com/arthelokyo/astrowind?utm_source=chatgpt.com "️ AstroWind: A free template using Astro 5 and Tailwind ..."
[3]: https://github.com/fuma-nama/fumadocs-ui-template?utm_source=chatgpt.com "fuma-nama/fumadocs-ui-template"
[4]: https://www.npmjs.com/package/fumadocs-openapi?utm_source=chatgpt.com "fumadocs-openapi"
