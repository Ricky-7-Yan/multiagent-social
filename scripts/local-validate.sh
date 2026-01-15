#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
echo "Starting services..."
docker compose -f "$ROOT/deployments/docker/docker-compose.yml" up -d --build
echo "Waiting for DB..."
sleep 6
export PG_DSN="postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
echo "Applying migrations..."
psql "$PG_DSN" -f "$ROOT/migrations/001_init.sql"
echo "Running unit tests..."
go test ./... -v
echo "Tests passed"
echo "Run manual checks: health /admin /ws endpoints"
