#!/bin/bash
set -e

echo "🔍 Running pre-push checks..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Format check
echo -e "\n${YELLOW}1. Checking code formatting...${NC}"
if [ -n "$(gofmt -l .)" ]; then
    echo -e "${RED}❌ Code is not formatted. Run: make fmt${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Code formatting OK${NC}"

# Vet
echo -e "\n${YELLOW}2. Running go vet...${NC}"
VET_OUTPUT=$(go vet ./... 2>&1 | grep -v "quic-go" || true)
if [ -z "$VET_OUTPUT" ]; then
    echo -e "${GREEN}✅ Go vet passed${NC}"
else
    echo -e "${RED}❌ Go vet found issues${NC}"
    echo "$VET_OUTPUT"
    exit 1
fi

# Unit tests
echo -e "\n${YELLOW}3. Running unit tests...${NC}"
if ! go test -short ./... -v; then
    echo -e "${RED}❌ Unit tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Unit tests passed${NC}"

# Build check
echo -e "\n${YELLOW}4. Checking build...${NC}"
if ! go build -o /tmp/zeno-auth-test ./cmd/auth > /dev/null 2>&1; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
rm -f /tmp/zeno-auth-test
echo -e "${GREEN}✅ Build successful${NC}"

# Check for sensitive data
echo -e "\n${YELLOW}5. Checking for sensitive data...${NC}"
if git diff --cached --name-only | xargs grep -i "password\|secret\|api_key" | grep -v "test\|example\|placeholder" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Warning: Possible sensitive data detected${NC}"
    echo "Please review your changes carefully"
fi

echo -e "\n${GREEN}✅ All pre-push checks passed!${NC}"
echo -e "${GREEN}🚀 Ready to push${NC}"
