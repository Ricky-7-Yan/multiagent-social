#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
echo "Starting docker compose..."
docker compose -f "$ROOT_DIR/deployments/docker/docker-compose.yml" up -d --build
echo "Waiting for postgres..."
sleep 5
export PG_DSN="postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
echo "Running migrations..."
psql "$PG_DSN" -f "$ROOT_DIR/migrations/001_init.sql"
echo "Running simple e2e"
# create conversation
CONV_ID=$(curl -s -X POST http://localhost:8080/api/v1/conversations)
echo "conversation id: $CONV_ID"
sleep 1
curl -s -X POST --data "hello from e2e" "http://localhost:8080/api/v1/conversations/$CONV_ID/messages"
echo "posted message"
echo "Check messages in DB (last 5):"
psql "$PG_DSN" -c "SELECT id, content, created_at FROM messages WHERE conversation_id = '$CONV_ID' ORDER BY created_at DESC LIMIT 5;"
echo "e2e done"
