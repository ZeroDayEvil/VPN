#!/bin/bash

# Build script for VPN Client
# This script builds the Windows client from Linux/Mac

set -e

echo "================================================"
echo "  VPN One-Click Client - Build Script"
echo "================================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo "Go version:"
go version
echo ""

# Navigate to client directory
cd "$(dirname "$0")/client"

echo "Installing dependencies..."
go mod tidy
go mod download
echo ""

# Build for Windows
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o Client.exe
echo ""

# Build for current platform (for testing)
echo "Building for current platform..."
go build -o client
echo ""

# Check build results
if [ -f "Client.exe" ]; then
    SIZE=$(du -h Client.exe | cut -f1)
    echo "✓ Windows build successful: Client.exe ($SIZE)"
else
    echo "✗ Windows build failed"
    exit 1
fi

if [ -f "client" ]; then
    SIZE=$(du -h client | cut -f1)
    echo "✓ Local build successful: client ($SIZE)"
else
    echo "✗ Local build failed"
    exit 1
fi

echo ""
echo "================================================"
echo "  Build Complete!"
echo "================================================"
echo ""
echo "Windows executable: client/Client.exe"
echo "Local executable:   client/client"
echo ""
echo "Next steps:"
echo "1. Test the local build: cd client && ./client -no-ui"
echo "2. Distribute Client.exe to Windows users"
echo ""
