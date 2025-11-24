#!/bin/bash
# SendGrid Setup Script for Zeno Auth

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
SERVICE_ACCOUNT="zeno-auth-sa@${PROJECT_ID}.iam.gserviceaccount.com"

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  SendGrid Setup for Zeno Auth                         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if API key is provided
if [ -z "$1" ]; then
    echo -e "${YELLOW}Usage: $0 <SENDGRID_API_KEY>${NC}"
    echo ""
    echo "Steps to get SendGrid API key:"
    echo "  1. Go to https://sendgrid.com/"
    echo "  2. Sign up for free account"
    echo "  3. Go to Settings → API Keys"
    echo "  4. Create API Key with 'Mail Send' permission"
    echo "  5. Copy the API key"
    echo ""
    echo "Then run:"
    echo "  $0 SG.your-api-key-here"
    echo ""
    exit 1
fi

SENDGRID_API_KEY="$1"

echo -e "${YELLOW}Step 1: Creating SendGrid API key secret...${NC}"

if gcloud secrets describe zeno-auth-sendgrid-api-key &>/dev/null; then
    echo -e "${GREEN}✅ Secret already exists, updating...${NC}"
    echo -n "$SENDGRID_API_KEY" | gcloud secrets versions add zeno-auth-sendgrid-api-key --data-file=-
else
    echo -e "${BLUE}Creating new secret...${NC}"
    echo -n "$SENDGRID_API_KEY" | gcloud secrets create zeno-auth-sendgrid-api-key \
        --data-file=- \
        --replication-policy=automatic
fi

echo -e "${GREEN}✅ Secret created/updated${NC}"
echo ""

echo -e "${YELLOW}Step 2: Granting access to service account...${NC}"

gcloud secrets add-iam-policy-binding zeno-auth-sendgrid-api-key \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/secretmanager.secretAccessor" \
    --quiet

echo -e "${GREEN}✅ Access granted${NC}"
echo ""

echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ SendGrid Setup Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Next steps:"
echo "  1. Verify sender email in SendGrid dashboard"
echo "  2. Deploy application: ./deploy/gcp-deploy.sh"
echo "  3. Test email sending"
echo ""
echo "Documentation: docs/EMAIL_SETUP.md"
echo ""
