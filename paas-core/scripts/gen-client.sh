#!/bin/bash
set -e

echo "=== Generating API Client ==="

# Step 1: Export OpenAPI spec from backend
echo "→ Generating OpenAPI spec..."
cd "$(dirname "$0")/../apps/api"
swag init -g cmd/server/main.go -o docs --outputTypes json
echo "  ✓ OpenAPI spec at apps/api/docs/openapi.json"

# Step 2: Generate TypeScript client via Orval
echo "→ Generating TypeScript client..."
cd "$(dirname "$0")/../packages/api-client"
pnpm generate
echo "  ✓ API client generated at packages/api-client/src/"

echo "=== Done ==="
