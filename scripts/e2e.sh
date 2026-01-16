#!/usr/bin/env bash
set -euo pipefail
echo "Running simple e2e test..."
export PG_DSN="postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"

# Wait for app to be ready
echo "Waiting for app to be ready..."
timeout 60 bash -c 'until curl -f http://localhost:8080/health; do sleep 2; done' || exit 1

echo "Creating conversation..."
CONV_ID=$(curl -s -X POST http://localhost:8080/api/v1/conversations)
echo "conversation id: $CONV_ID"

if [ -z "$CONV_ID" ] || [ "$CONV_ID" = "null" ]; then
  echo "Failed to create conversation"
  exit 1
fi

sleep 1
echo "Posting message..."
curl -s -X POST --data "hello from e2e" "http://localhost:8080/api/v1/conversations/$CONV_ID/messages"
echo "posted message"

echo "Checking messages in DB..."
psql "$PG_DSN" -c "SELECT id, content, created_at FROM messages WHERE conversation_id = '$CONV_ID' ORDER BY created_at DESC LIMIT 5;"

echo "E2E test completed successfully"
