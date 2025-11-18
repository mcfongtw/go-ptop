#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME=go-ptop-poc
CONTAINER_NAME=go-ptop-poc-run

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

cd "$REPO_ROOT"

echo "[1/2] Building Docker image $IMAGE_NAME..."
docker build --no-cache -t "$IMAGE_NAME" -f build/Dockerfile .

echo "[2/2] Running container to exercise go-ptop against sample JVM..."
docker run --rm --name "$CONTAINER_NAME" "$IMAGE_NAME"
