#!/bin/bash
# Pre-Deployment Validation Script
# Checks all requirements before deploying to GCP

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
REGION="${REGION:-europe-west3}"
INSTANCE_ID="${INSTANCE_ID:-zeno-auth-db-dev}"
SERVICE_ACCOUNT="zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com"

ERRORS=0
WARNINGS=0

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🔍 Pre-Deployment Validation${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Check 1: gcloud CLI
echo -n "Checking gcloud CLI... "
if command -v gcloud &> /dev/null; then
    echo -e "${GREEN}✅${NC}"
else
    echo -e "${RED}❌ gcloud CLI not found${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check 2: Docker
echo -n "Checking Docker... "
if command -v docker &> /dev/null && docker info &> /dev/null; then
    echo -e "${GREEN}✅${NC}"
else
    echo -e "${RED}❌ Docker not running${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check 3: Project access
echo -n "Checking GCP project access... "
if gcloud config get-value project &> /dev/null; then
    CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
    if [ "$CURRENT_PROJECT" = "$PROJECT_ID" ]; then
        echo -e "${GREEN}✅ ($CURRENT_PROJECT)${NC}"
    else
        echo -e "${YELLOW}⚠️  Current: $CURRENT_PROJECT, Expected: $PROJECT_ID${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${RED}❌ Not authenticated${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check 4: Cloud SQL instance
echo -n "Checking Cloud SQL instance... "
INSTANCE_STATE=$(gcloud sql instances describe "$INSTANCE_ID" --format="value(state)" 2>/dev/null || echo "NOT_FOUND")
if [ "$INSTANCE_STATE" = "RUNNABLE" ]; then
    echo -e "${GREEN}✅ RUNNABLE${NC}"
elif [ "$INSTANCE_STATE" = "NOT_FOUND" ]; then
    echo -e "${RED}❌ Instance not found${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${RED}❌ State: $INSTANCE_STATE${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check 5: Service Account
echo -n "Checking Service Account... "
if gcloud iam service-accounts describe "$SERVICE_ACCOUNT" &> /dev/null; then
    echo -e "${GREEN}✅${NC}"
else
    echo -e "${RED}❌ Service Account not found${NC}"
    echo "  Run: ./gcp-setup-secrets.sh"
    ERRORS=$((ERRORS + 1))
fi

# Check 6: IAM Roles
echo "Checking IAM roles..."
REQUIRED_ROLES=(
    "roles/cloudsql.client"
    "roles/secretmanager.secretAccessor"
    "roles/logging.logWriter"
)

for role in "${REQUIRED_ROLES[@]}"; do
    echo -n "  - $role... "
    if gcloud projects get-iam-policy "$PROJECT_ID" \
        --flatten="bindings[].members" \
        --filter="bindings.members:serviceAccount:$SERVICE_ACCOUNT AND bindings.role:$role" \
        --format="value(bindings.role)" 2>/dev/null | grep -q "$role"; then
        echo -e "${GREEN}✅${NC}"
    else
        echo -e "${RED}❌${NC}"
        ERRORS=$((ERRORS + 1))
    fi
done

# Check 7: DATABASE_URL Secret
echo -n "Checking DATABASE_URL secret... "
if gcloud secrets describe zeno-auth-database-url &> /dev/null; then
    VERSION=$(gcloud secrets versions list zeno-auth-database-url --format="value(name)" --limit=1)
    if [ -n "$VERSION" ]; then
        echo -e "${GREEN}✅ (version: $VERSION)${NC}"
    else
        echo -e "${RED}❌ No versions${NC}"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${RED}❌ Secret not found${NC}"
    echo "  Run: ./gcp-setup-secrets.sh"
    ERRORS=$((ERRORS + 1))
fi

# Check 8: JWT_PRIVATE_KEY Secret
echo -n "Checking JWT_PRIVATE_KEY secret... "
if gcloud secrets describe zeno-auth-jwt-private-key &> /dev/null; then
    VERSION=$(gcloud secrets versions list zeno-auth-jwt-private-key --format="value(name)" --limit=1)
    if [ -n "$VERSION" ]; then
        echo -e "${GREEN}✅ (version: $VERSION)${NC}"
    else
        echo -e "${YELLOW}⚠️  No versions${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Secret not found (will use embedded key)${NC}"
    echo "  Recommended: ./gcp-setup-secrets.sh"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 9: Artifact Registry
echo -n "Checking Artifact Registry... "
REPO_NAME="${REPO_NAME:-zeno-auth}"
if gcloud artifacts repositories describe "$REPO_NAME" --location="$REGION" &> /dev/null; then
    echo -e "${GREEN}✅${NC}"
else
    echo -e "${YELLOW}⚠️  Repository will be created during deployment${NC}"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 10: Local tests
echo -n "Checking local tests... "
cd "$(dirname "$0")/.."
if go test ./... -short &> /dev/null; then
    echo -e "${GREEN}✅${NC}"
else
    echo -e "${YELLOW}⚠️  Some tests failed${NC}"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 11: Code quality
echo -n "Checking code formatting... "
if [ -z "$(gofmt -l . 2>/dev/null | grep -v vendor)" ]; then
    echo -e "${GREEN}✅${NC}"
else
    echo -e "${YELLOW}⚠️  Code needs formatting (run: go fmt ./...)${NC}"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 12: No secrets in code
echo -n "Checking for secrets in code... "
if git grep -i "BEGIN RSA PRIVATE KEY" &> /dev/null || \
   git grep -i "BEGIN PRIVATE KEY" &> /dev/null; then
    echo -e "${RED}❌ Private keys found in git!${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✅${NC}"
fi

# Check 13: Dockerfile
echo -n "Checking Dockerfile... "
if [ -f "Dockerfile" ]; then
    if grep -q "migrate" Dockerfile; then
        echo -e "${GREEN}✅ (migrate included)${NC}"
    else
        echo -e "${YELLOW}⚠️  migrate not found in Dockerfile${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${RED}❌ Dockerfile not found${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Summary
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 Validation Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed! Ready to deploy.${NC}"
    echo ""
    echo "Run: ./deploy/gcp-deploy.sh"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  $WARNINGS warning(s) found${NC}"
    echo -e "${GREEN}✅ No critical errors. You can proceed with deployment.${NC}"
    echo ""
    echo "Run: ./deploy/gcp-deploy.sh"
    exit 0
else
    echo -e "${RED}❌ $ERRORS error(s) found${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}⚠️  $WARNINGS warning(s) found${NC}"
    fi
    echo ""
    echo "Please fix the errors before deploying:"
    echo "  1. Run: ./deploy/gcp-setup-secrets.sh"
    echo "  2. Verify Cloud SQL instance is running"
    echo "  3. Check IAM roles and secrets"
    echo ""
    exit 1
fi
