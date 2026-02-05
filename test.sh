#!/bin/bash

# Test script for VPN project
# Tests both backend and client components

set -e

echo "================================================"
echo "  VPN One-Click Client - Test Suite"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(dirname "$0")"
cd "$PROJECT_ROOT"

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

test_pass() {
    echo -e "${GREEN}✓${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

test_fail() {
    echo -e "${RED}✗${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

echo "1. Testing Backend API..."
echo "----------------------------------------"

# Check if Python3 is installed
if command -v python3 &> /dev/null; then
    test_pass "Python3 is installed"
else
    test_fail "Python3 is not installed"
fi

# Check backend files
if [ -f "backend/main.py" ]; then
    test_pass "Backend main.py exists"
else
    test_fail "Backend main.py not found"
fi

if [ -f "backend/requirements.txt" ]; then
    test_pass "Backend requirements.txt exists"
else
    test_fail "Backend requirements.txt not found"
fi

# Try to import FastAPI
if python3 -c "import fastapi" 2>/dev/null; then
    test_pass "FastAPI is installed"
else
    echo -e "${YELLOW}⚠${NC} FastAPI not installed (run: pip3 install -r backend/requirements.txt)"
fi

# Check Python syntax
if python3 -m py_compile backend/main.py 2>/dev/null; then
    test_pass "Backend Python syntax is valid"
else
    test_fail "Backend Python syntax error"
fi

echo ""
echo "2. Testing Client Build..."
echo "----------------------------------------"

# Check if Go is installed
if command -v go &> /dev/null; then
    test_pass "Go is installed ($(go version))"
else
    test_fail "Go is not installed"
fi

# Check client files
if [ -f "client/main.go" ]; then
    test_pass "Client main.go exists"
else
    test_fail "Client main.go not found"
fi

if [ -f "client/go.mod" ]; then
    test_pass "Client go.mod exists"
else
    test_fail "Client go.mod not found"
fi

# Test Go build
cd client
if go build -o test_client 2>/dev/null; then
    test_pass "Client builds successfully"
    rm -f test_client
else
    test_fail "Client build failed"
fi
cd ..

echo ""
echo "3. Testing Configuration..."
echo "----------------------------------------"

# Check config files
if [ -f "config/manifest.json" ]; then
    test_pass "Manifest file exists"
else
    test_fail "Manifest file not found"
fi

if [ -f "config/config.json" ]; then
    test_pass "Config file exists"
else
    test_fail "Config file not found"
fi

# Validate JSON syntax
if python3 -c "import json; json.load(open('config/manifest.json'))" 2>/dev/null; then
    test_pass "Manifest JSON is valid"
else
    test_fail "Manifest JSON is invalid"
fi

if python3 -c "import json; json.load(open('config/config.json'))" 2>/dev/null; then
    test_pass "Config JSON is valid"
else
    test_fail "Config JSON is invalid"
fi

# Verify SHA256 in manifest matches config
MANIFEST_SHA=$(python3 -c "import json; print(json.load(open('config/manifest.json'))['sha256'])")
ACTUAL_SHA=$(sha256sum config/config.json | cut -d' ' -f1)
if [ "$MANIFEST_SHA" = "$ACTUAL_SHA" ]; then
    test_pass "Config SHA256 matches manifest"
else
    echo -e "${YELLOW}⚠${NC} Config SHA256 mismatch (run: sha256sum config/config.json and update manifest)"
fi

echo ""
echo "4. Testing Documentation..."
echo "----------------------------------------"

DOCS=("README.md" "docs/TECHNICAL.md" "docs/USER_GUIDE.md" "docs/DEPLOYMENT.md" "docs/PROXY_NODE_SETUP.md")
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        test_pass "$doc exists"
    else
        test_fail "$doc not found"
    fi
done

echo ""
echo "================================================"
echo "  Test Results"
echo "================================================"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Start backend: cd backend && python3 main.py"
    echo "2. Test client: cd client && go run main.go -no-ui"
    echo "3. Build for production: ./build.sh"
    exit 0
else
    echo -e "${RED}Some tests failed. Please fix the issues above.${NC}"
    exit 1
fi
