#!/bin/bash
set -e

echo "=== Testing Go Build Locally ==="
echo ""

cd src

echo "1. Checking Go version..."
go version

echo ""
echo "2. Cleaning go cache..."
go clean -cache -modcache -testcache 2>/dev/null || true

echo ""
echo "3. Running go mod tidy..."
go mod tidy

echo ""
echo "4. Running go mod download..."
go mod download

echo ""
echo "5. Running go mod verify..."
go mod verify

echo ""
echo "6. Testing build (this will show actual error)..."
go build -v -o whatsapp-test 2>&1

echo ""
echo "=== Build SUCCESS! ==="
echo "Binary created: src/whatsapp-test"
