#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-multiagent-social}"
REGISTRY="${REGISTRY:-}"
TAG="${TAG:-$(git rev-parse --short HEAD)}"

echo "Building frontend..."
cd "$ROOT_DIR/web/react-app"
npm ci
npm run build

echo "Building Docker image..."
cd "$ROOT_DIR"
docker build -t "${IMAGE_NAME}:${TAG}" .

if [ -n "$REGISTRY" ]; then
  FULL="${REGISTRY}/${IMAGE_NAME}:${TAG}"
  docker tag "${IMAGE_NAME}:${TAG}" "${FULL}"
  echo "Pushing ${FULL}..."
  docker push "${FULL}"
else
  echo "REGISTRY not set, skipping push. Built image: ${IMAGE_NAME}:${TAG}"
fi

