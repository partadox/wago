#!/bin/bash
set -e

echo "=== Testing Docker Build Locally ==="
echo ""
echo "This will build the Docker image and show any compilation errors"
echo ""

docker build -f docker/golang.Dockerfile -t wago-test:latest . 2>&1 | tee docker-build.log

echo ""
echo "=== Build logs saved to: docker-build.log ==="
echo ""
echo "If build succeeded, you can run:"
echo "  docker run --rm wago-test:latest --help"
