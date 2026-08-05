#!/bin/bash

set -euo pipefail

VERSION_FILE="VERSION"
VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')

if [ -z "$VERSION" ]; then
    echo "Error: VERSION file is empty"
    exit 1
fi

GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS="-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE} -w -s"

echo "Building with version: ${VERSION}"
echo "Git commit: ${GIT_COMMIT}"
echo "Build date: ${BUILD_DATE}"

export LDFLAGS
export VERSION
export GIT_COMMIT
export BUILD_DATE
