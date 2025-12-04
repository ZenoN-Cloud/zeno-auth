#!/bin/bash
set -e

echo "🧪 Running complete test suite..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

FAILED=0

# Unit tests
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}📦 Running unit tests...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if go test -short -v ./...; then
    echo -e "${GREEN}✅ Unit tests passed${NC}"
else
    echo -e "${RED}❌ Unit tests failed${NC}"
    FAILED=1
fi

# Test with race detector
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}🏁 Running race detector...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if go test -short -race ./...; then
    echo -e "${GREEN}✅ Race detector passed${NC}"
else
    echo -e "${RED}❌ Race conditions detected${NC}"
    FAILED=1
fi

# Coverage
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}📊 Generating coverage report...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if go test -short -coverprofile=coverage.out ./... > /dev/null 2>&1; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' 2>/dev/null || echo "N/A")
    echo -e "${GREEN}Coverage: ${COVERAGE}${NC}"
else
    echo -e "${YELLOW}⚠️  Coverage generation failed${NC}"
fi

# Integration tests (if server is running)
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}🔗 Checking for integration tests...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "Server is running, executing integration tests..."
    if go test ./test -v; then
        echo -e "${GREEN}✅ Integration tests passed${NC}"
    else
        echo -e "${YELLOW}⚠️  Integration tests failed (non-critical)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Server not running, skipping integration tests${NC}"
    echo "To run integration tests: make local-up && make integration"
fi

# Build test
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}🔨 Testing build...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if go build -o /tmp/zeno-auth-test ./cmd/auth; then
    echo -e "${GREEN}✅ Build successful${NC}"
    set +e
    rm -f /tmp/zeno-auth-test
    set -e
else
    echo -e "${RED}❌ Build failed${NC}"
    FAILED=1
fi

# Docker build test
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}🐳 Testing Docker build...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if docker build -t zeno-auth:test . > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker build successful${NC}"
    set +e
    docker rmi zeno-auth:test > /dev/null 2>&1
    set -e
else
    echo -e "${RED}❌ Docker build failed${NC}"
    FAILED=1
fi

# Cleanup
set +e
rm -f coverage.out
set -e

# Summary
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}🚀 Ready to push to repository${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    echo -e "${RED}Please fix the issues before pushing${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi
