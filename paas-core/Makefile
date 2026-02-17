.PHONY: dev api-dev web-dev api-test web-test openapi gen-client lint

# Local development (requires Docker)
dev:
	cd deploy/docker && docker compose up --build

dev-down:
	cd deploy/docker && docker compose down

# API development (without Docker)
api-dev:
	cd apps/api && go run ./cmd/server

api-test:
	cd apps/api && go test ./...

api-lint:
	cd apps/api && go vet ./...

# Web development
web-dev:
	cd apps/web && pnpm dev

web-test:
	cd apps/web && pnpm test

web-lint:
	cd apps/web && pnpm lint

# OpenAPI
openapi:
	cd apps/api && swag init -g cmd/server/main.go -o docs --outputTypes json
	@echo "OpenAPI spec generated at apps/api/docs/openapi.json"

# Client generation (Orval)
gen-client:
	cd packages/api-client && pnpm generate
	@echo "TypeScript API client generated"

# Lint all
lint: api-lint web-lint
	@echo "All lint checks passed"

# Database
db-migrate:
	@echo "Run migrations via docker-entrypoint-initdb.d (automatic on first run)"
	@echo "For manual: psql -h localhost -U paas -d paas -f apps/api/migrations/001_init.sql"

# Docker builds
build-api:
	docker build -f deploy/docker/Dockerfile.api -t paas-api .

build-web:
	docker build -f deploy/docker/Dockerfile.web -t paas-web .

build: build-api build-web
