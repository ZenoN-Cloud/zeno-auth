#!/bin/bash
# GitLab CI/CD Variables Setup for Zeno-CY

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🔧 GitLab CI/CD Variables Setup${NC}"

PROJECT_PATH="maxim.viazov/zeno-cy/zeno-auth"

if ! command -v glab &> /dev/null; then
    echo -e "${RED}❌ GitLab CLI (glab) not found${NC}"
    echo "Install: https://gitlab.com/gitlab-org/cli"
    exit 1
fi

echo -e "${YELLOW}📋 Setting up CI/CD variables...${NC}"

# GCP Service Account Key (DEV)
if [ -f "$HOME/.config/gcloud/zeno-cy-dev-sa.json" ]; then
    if ! GCP_KEY_DEV=$(base64 -i "$HOME/.config/gcloud/zeno-cy-dev-sa.json"); then
        echo -e "${RED}❌ Failed to encode GCP service account key${NC}"
        exit 1
    fi
    if ! glab variable set GCP_SERVICE_ACCOUNT_KEY "$GCP_KEY_DEV" \
        --scope=* --masked --project="$PROJECT_PATH"; then
        echo -e "${RED}❌ Failed to set GCP_SERVICE_ACCOUNT_KEY variable${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ GCP_SERVICE_ACCOUNT_KEY set${NC}"
else
    echo -e "${RED}❌ GCP service account key not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Setup complete!${NC}"