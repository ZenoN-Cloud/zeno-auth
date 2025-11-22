#!/bin/bash
# GCP Cloud Run Deployment Script for Zeno Auth
# Version: 1.1.0

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
REGION="${REGION:-europe-west3}"
INSTANCE_ID="${INSTANCE_ID:-zeno-auth-db-dev}"
DB_NAME="${DB_NAME:-zeno_auth}"
DB_USER="${DB_USER:-zeno_auth_app}"
INSTANCE_CONNECTION_NAME="${INSTANCE_CONNECTION_NAME:-$PROJECT_ID:$REGION:$INSTANCE_ID}"
SERVICE_NAME="${SERVICE_NAME:-zeno-auth-dev}"
REPO_NAME="${REPO_NAME:-zeno-auth}"

echo -e "${GREEN}🚀 Zeno Auth - GCP Cloud Run Deployment${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo "Service: $SERVICE_NAME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Step 1: Check prerequisites
echo -e "${YELLOW}📋 Step 1: Checking prerequisites...${NC}"

if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI not found. Please install it first.${NC}"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker not found. Please install it first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites OK${NC}"
echo ""

# Step 2: Set GCP project
echo -e "${YELLOW}📋 Step 2: Setting GCP project...${NC}"
gcloud config set project "$PROJECT_ID"
echo -e "${GREEN}✅ Project set${NC}"
echo ""

# Step 3: Check Cloud SQL instance
echo -e "${YELLOW}📋 Step 3: Checking Cloud SQL instance...${NC}"
INSTANCE_STATE=$(gcloud sql instances describe "$INSTANCE_ID" --format="value(state)" 2>/dev/null || echo "NOT_FOUND")

if [ "$INSTANCE_STATE" != "RUNNABLE" ]; then
    echo -e "${RED}❌ Cloud SQL instance is not RUNNABLE (state: $INSTANCE_STATE)${NC}"
    echo "Please ensure the database is ready before deploying."
    exit 1
fi

echo -e "${GREEN}✅ Cloud SQL instance is RUNNABLE${NC}"
echo ""

# Step 4: Check Secret Manager
echo -e "${YELLOW}📋 Step 4: Checking Secret Manager...${NC}"
if ! gcloud secrets describe zeno-auth-database-url &> /dev/null; then
    echo -e "${RED}❌ Secret 'zeno-auth-database-url' not found${NC}"
    echo "Please create the secret first:"
    echo "  gcloud secrets create zeno-auth-database-url --data-file=-"
    exit 1
fi

SECRET_VERSION=$(gcloud secrets versions list zeno-auth-database-url --format="value(name)" --limit=1)
if [ -z "$SECRET_VERSION" ]; then
    echo -e "${RED}❌ No secret versions found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Secret exists (version: $SECRET_VERSION)${NC}"
echo ""

# Step 5: Create Artifact Registry repository (if not exists)
echo -e "${YELLOW}📋 Step 5: Checking Artifact Registry...${NC}"
if ! gcloud artifacts repositories describe "$REPO_NAME" --location="$REGION" &> /dev/null; then
    echo "Creating Artifact Registry repository..."
    gcloud artifacts repositories create "$REPO_NAME" \
        --repository-format=docker \
        --location="$REGION" \
        --description="Zeno Auth Docker images"
    echo -e "${GREEN}✅ Repository created${NC}"
else
    echo -e "${GREEN}✅ Repository exists${NC}"
fi
echo ""

# Step 6: Build and push Docker image
echo -e "${YELLOW}📋 Step 6: Building and pushing Docker image...${NC}"
IMAGE="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO_NAME/zeno-auth:$(date +%Y%m%d-%H%M%S)"
IMAGE_LATEST="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO_NAME/zeno-auth:latest"

echo "Building image: $IMAGE"
gcloud builds submit --tag "$IMAGE" --tag "$IMAGE_LATEST"

echo -e "${GREEN}✅ Image built and pushed${NC}"
echo ""

# Step 7: Deploy to Cloud Run
echo -e "${YELLOW}📋 Step 7: Deploying to Cloud Run...${NC}"

# Check if service exists
if gcloud run services describe "$SERVICE_NAME" --region="$REGION" &> /dev/null; then
    echo "Updating existing service..."
    gcloud run deploy "$SERVICE_NAME" \
        --image="$IMAGE" \
        --region="$REGION" \
        --platform=managed \
        --add-cloudsql-instances="$INSTANCE_CONNECTION_NAME" \
        --set-secrets=DATABASE_URL=zeno-auth-database-url:latest \
        --port=8080 \
        --memory=512Mi \
        --cpu=1 \
        --timeout=300 \
        --max-instances=10 \
        --min-instances=0 \
        --allow-unauthenticated
else
    echo "Creating new service..."
    gcloud run deploy "$SERVICE_NAME" \
        --image="$IMAGE" \
        --region="$REGION" \
        --platform=managed \
        --add-cloudsql-instances="$INSTANCE_CONNECTION_NAME" \
        --set-secrets=DATABASE_URL=zeno-auth-database-url:latest \
        --allow-unauthenticated \
        --port=8080 \
        --memory=512Mi \
        --cpu=1 \
        --timeout=300 \
        --max-instances=10 \
        --min-instances=0
fi

echo -e "${GREEN}✅ Service deployed${NC}"
echo ""

# Step 8: Get service URL
echo -e "${YELLOW}📋 Step 8: Getting service URL...${NC}"
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region="$REGION" --format="value(status.url)")

echo -e "${GREEN}✅ Service URL: $SERVICE_URL${NC}"
echo ""

# Step 9: Health check
echo -e "${YELLOW}📋 Step 9: Running health check...${NC}"
sleep 5

if curl -s -f "$SERVICE_URL/health" > /dev/null; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    curl -s "$SERVICE_URL/health" | jq .
else
    echo -e "${RED}❌ Health check failed${NC}"
    echo "Check logs: gcloud logs read $SERVICE_NAME --region=$REGION --limit=50"
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}🎉 Deployment Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Service URL: $SERVICE_URL"
echo "Image: $IMAGE"
echo ""
echo "Next steps:"
echo "  1. Test API: curl $SERVICE_URL/health"
echo "  2. View logs: gcloud logs read $SERVICE_NAME --region=$REGION --limit=50"
echo "  3. Monitor: https://console.cloud.google.com/run/detail/$REGION/$SERVICE_NAME"
echo ""
