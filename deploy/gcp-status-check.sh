#!/bin/bash
# GCP Infrastructure Status Check Script

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
DB_NAME="${DB_NAME:-zeno_auth}"
DB_USER="${DB_USER:-zeno_auth_app}"
SERVICE_NAME="${SERVICE_NAME:-zeno-auth-dev}"
SERVICE_ACCOUNT_EMAIL="zeno-auth-sa@${PROJECT_ID}.iam.gserviceaccount.com"

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Zeno Auth - GCP Infrastructure Status Check          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Set project
gcloud config set project "$PROJECT_ID" --quiet

# Check Cloud SQL Instance
echo -e "${YELLOW}━━━ Cloud SQL Instance ━━━${NC}"
if gcloud sql instances describe "$INSTANCE_ID" &>/dev/null; then
    STATE=$(gcloud sql instances describe "$INSTANCE_ID" --format="value(state)")
    if [ "$STATE" = "RUNNABLE" ]; then
        echo -e "  Status: ${GREEN}✅ RUNNABLE${NC}"
    else
        echo -e "  Status: ${YELLOW}⚠️  $STATE${NC}"
    fi
    
    IP=$(gcloud sql instances describe "$INSTANCE_ID" --format="value(ipAddresses[0].ipAddress)")
    echo -e "  Instance: ${BLUE}$INSTANCE_ID${NC}"
    echo -e "  IP: ${BLUE}$IP${NC}"
else
    echo -e "  Status: ${RED}❌ NOT FOUND${NC}"
fi
echo ""

# Check Database
echo -e "${YELLOW}━━━ Database ━━━${NC}"
if gcloud sql databases describe "$DB_NAME" --instance="$INSTANCE_ID" &>/dev/null; then
    echo -e "  Status: ${GREEN}✅ EXISTS${NC}"
    echo -e "  Name: ${BLUE}$DB_NAME${NC}"
else
    echo -e "  Status: ${RED}❌ NOT FOUND${NC}"
fi
echo ""

# Check Database User
echo -e "${YELLOW}━━━ Database User ━━━${NC}"
if gcloud sql users describe "$DB_USER" --instance="$INSTANCE_ID" &>/dev/null; then
    echo -e "  Status: ${GREEN}✅ EXISTS${NC}"
    echo -e "  User: ${BLUE}$DB_USER${NC}"
else
    echo -e "  Status: ${RED}❌ NOT FOUND${NC}"
fi
echo ""

# Check Secrets
echo -e "${YELLOW}━━━ Secret Manager ━━━${NC}"

if gcloud secrets describe zeno-auth-database-url &>/dev/null; then
    VERSION=$(gcloud secrets versions list zeno-auth-database-url --format="value(name)" --limit=1)
    echo -e "  DATABASE_URL: ${GREEN}✅ EXISTS${NC} (version: $VERSION)"
else
    echo -e "  DATABASE_URL: ${RED}❌ NOT FOUND${NC}"
fi

if gcloud secrets describe zeno-auth-jwt-private-key &>/dev/null; then
    VERSION=$(gcloud secrets versions list zeno-auth-jwt-private-key --format="value(name)" --limit=1)
    echo -e "  JWT_PRIVATE_KEY: ${GREEN}✅ EXISTS${NC} (version: $VERSION)"
else
    echo -e "  JWT_PRIVATE_KEY: ${RED}❌ NOT FOUND${NC}"
fi
echo ""

# Check Service Account
echo -e "${YELLOW}━━━ Service Account ━━━${NC}"
if gcloud iam service-accounts describe "$SERVICE_ACCOUNT_EMAIL" &>/dev/null; then
    echo -e "  Status: ${GREEN}✅ EXISTS${NC}"
    echo -e "  Email: ${BLUE}$SERVICE_ACCOUNT_EMAIL${NC}"
    
    # Check roles
    echo -e "  Roles:"
    ROLES=$(gcloud projects get-iam-policy "$PROJECT_ID" \
        --flatten="bindings[].members" \
        --filter="bindings.members:serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
        --format="value(bindings.role)")
    
    if [ -n "$ROLES" ]; then
        echo "$ROLES" | while read -r role; do
            echo -e "    ${GREEN}✓${NC} $role"
        done
    else
        echo -e "    ${YELLOW}⚠️  No roles assigned${NC}"
    fi
else
    echo -e "  Status: ${RED}❌ NOT FOUND${NC}"
fi
echo ""

# Check Artifact Registry
echo -e "${YELLOW}━━━ Artifact Registry ━━━${NC}"
if gcloud artifacts repositories describe zeno-auth --location="$REGION" &>/dev/null; then
    echo -e "  Status: ${GREEN}✅ EXISTS${NC}"
    echo -e "  Location: ${BLUE}$REGION-docker.pkg.dev/$PROJECT_ID/zeno-auth${NC}"
    
    # Count images
    IMAGE_COUNT=$(gcloud artifacts docker images list "$REGION-docker.pkg.dev/$PROJECT_ID/zeno-auth" --format="value(package)" 2>/dev/null | wc -l)
    echo -e "  Images: ${BLUE}$IMAGE_COUNT${NC}"
else
    echo -e "  Status: ${RED}❌ NOT FOUND${NC}"
fi
echo ""

# Check Cloud Run Service
echo -e "${YELLOW}━━━ Cloud Run Service ━━━${NC}"
if gcloud run services describe "$SERVICE_NAME" --region="$REGION" &>/dev/null; then
    echo -e "  Status: ${GREEN}✅ DEPLOYED${NC}"
    
    URL=$(gcloud run services describe "$SERVICE_NAME" --region="$REGION" --format="value(status.url)")
    echo -e "  URL: ${BLUE}$URL${NC}"
    
    REVISION=$(gcloud run services describe "$SERVICE_NAME" --region="$REGION" --format="value(status.latestReadyRevisionName)")
    echo -e "  Latest Revision: ${BLUE}$REVISION${NC}"
    
    # Test health endpoint
    echo -e "  Health Check:"
    if curl -s -f "$URL/health" > /dev/null 2>&1; then
        echo -e "    ${GREEN}✅ HEALTHY${NC}"
        HEALTH=$(curl -s "$URL/health" | jq -r '.status' 2>/dev/null || echo "unknown")
        echo -e "    Status: ${BLUE}$HEALTH${NC}"
    else
        echo -e "    ${RED}❌ UNHEALTHY${NC}"
    fi
else
    echo -e "  Status: ${YELLOW}⚠️  NOT DEPLOYED${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Count ready components
READY=0
TOTAL=7

gcloud sql instances describe "$INSTANCE_ID" &>/dev/null && ((READY++))
gcloud sql databases describe "$DB_NAME" --instance="$INSTANCE_ID" &>/dev/null && ((READY++))
gcloud secrets describe zeno-auth-database-url &>/dev/null && ((READY++))
gcloud secrets describe zeno-auth-jwt-private-key &>/dev/null && ((READY++))
gcloud iam service-accounts describe "$SERVICE_ACCOUNT_EMAIL" &>/dev/null && ((READY++))
gcloud artifacts repositories describe zeno-auth --location="$REGION" &>/dev/null && ((READY++))
gcloud run services describe "$SERVICE_NAME" --region="$REGION" &>/dev/null && ((READY++))

PERCENT=$((READY * 100 / TOTAL))

echo -e "Infrastructure Ready: ${BLUE}$READY/$TOTAL${NC} (${BLUE}${PERCENT}%${NC})"

if [ $READY -eq $TOTAL ]; then
    echo -e "Status: ${GREEN}✅ FULLY DEPLOYED${NC}"
elif [ $READY -ge 6 ]; then
    echo -e "Status: ${YELLOW}⚠️  ALMOST READY (missing Cloud Run service)${NC}"
    echo -e "Next: Run ${BLUE}./deploy/gcp-deploy.sh${NC}"
elif [ $READY -ge 3 ]; then
    echo -e "Status: ${YELLOW}⚠️  PARTIALLY READY${NC}"
    echo -e "Next: Run ${BLUE}./deploy/gcp-setup-complete.sh${NC}"
else
    echo -e "Status: ${RED}❌ NOT READY${NC}"
    echo -e "Next: Run ${BLUE}./deploy/gcp-setup-complete.sh${NC}"
fi

echo ""
