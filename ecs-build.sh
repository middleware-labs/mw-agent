#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONTRIB_REPO="$SCRIPT_DIR/../opentelemetry-collector-contrib"
IMAGE="ghcr.io/middleware-labs/mw-host-agent"
TAG="dev-$(date +%s)"
ECS_CLUSTER="app-cluster"
ECS_SERVICE="mw-agent-daemon-service"

PACKAGES=(
  "receiver/awsecscontainermetricsreceiver"
  "internal/aws/ecsutil"
  "internal/common"
)

for pkg in "${PACKAGES[@]}"; do
  src="$CONTRIB_REPO/$pkg"
  dest="$SCRIPT_DIR/local-deps/$pkg"

  if [ ! -d "$src" ]; then
    echo "Error: $src not found"
    exit 1
  fi

  rm -rf "$dest"
  mkdir -p "$dest"
  cp -r "$src/"* "$dest/"
  echo "Copied $pkg"
done

echo "local-deps/ ready"
echo "Building $IMAGE:$TAG ..."

docker buildx build --platform linux/amd64 --progress=plain \
  -f Dockerfiles/DockerfileDebug \
  -t "$IMAGE:$TAG" . \
  --push

echo ""
echo "Pushed $IMAGE:$TAG"
echo "Update your ECS task definition to use this tag, then:"
echo "  aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --force-new-deployment --region us-east-1"
