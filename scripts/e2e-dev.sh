#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "Running devserver e2e..."

# create conversation
CONV_ID=$(curl -s -X POST http://localhost:8080/api/v1/conversations)
echo "conversation id: $CONV_ID"
sleep 1

echo "Posting message..."
curl -s -X POST --data "hello from dev e2e" "http://localhost:8080/api/v1/conversations/$CONV_ID/messages"
echo "posted message"

echo "Fetching messages..."
curl -s "http://localhost:8080/api/v1/conversations/$CONV_ID/messages" | jq || true

echo "devserver e2e completed"

