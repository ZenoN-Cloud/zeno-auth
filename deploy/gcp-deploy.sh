#!/bin/bash
# GCP Cloud Run Deployment Script for Zeno Auth
# Version: 1.2.0

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ─────────────────────────────────────────────
# Base configuration (DEV)
# ─────────────────────────────────────────────
PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
REGION="${REGION:-europe-west3}"

# Cloud SQL
INSTANCE_ID="${INSTANCE_ID:-zeno-auth-db-dev}"
DB_NAME="${DB_NAME:-zeno_auth}"
DB_USER="${DB_USER:-zeno_auth_app}"
INSTANCE_CONNECTION_NAME="${INSTANCE_CONNECTION_NAME:-$PROJECT_ID:$REGION:$INSTANCE_ID}"

# Cloud Run
SERVICE_NAME="${SERVICE_NAME:-zeno-auth-dev}"
REPO_NAME="${REPO_NAME:-zeno-auth}"
SERVICE_ACCOUNT="${SERVICE_ACCOUNT:-zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com}"

# App env
APP_ENV="${APP_ENV:-development}"
FRONTEND_BUCKET="${FRONTEND_BUCKET:-zeno-cy-frontend-dev-001}"
APP_BASE_URL="${APP_BASE_URL:-https://storage.googleapis.com/$FRONTEND_BUCKET}"

echo -e "${GREEN}🚀 Zeno Auth - GCP Cloud Run Deployment${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Project:            $PROJECT_ID"
echo "Region:             $REGION"
echo "Service:            $SERVICE_NAME"
echo "SQL instance:       $INSTANCE_ID"
echo "Connection name:    $INSTANCE_CONNECTION_NAME"
echo "App env:            $APP_ENV"
echo "Frontend base URL:  $APP_BASE_URL"
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
gcloud config set project "$PROJECT_ID" >/dev/null
echo -e "${GREEN}✅ Project set to $PROJECT_ID${NC}"
echo ""

# Step 3: Check Cloud SQL instance
echo -e "${YELLOW}📋 Step 3: Checking Cloud SQL instance...${NC}"
INSTANCE_STATE=$(gcloud sql instances describe "$INSTANCE_ID" \
  --format="value(state)" 2>/dev/null || echo "NOT_FOUND")

INSTANCE_REGION=$(gcloud sql instances describe "$INSTANCE_ID" \
  --format="value(region)" 2>/dev/null || echo "UNKNOWN")

echo "Instance region: $INSTANCE_REGION"

if [ "$INSTANCE_REGION" != "$REGION" ]; then
    echo -e "${YELLOW}⚠️  WARNING: Cloud SQL instance region is '$INSTANCE_REGION', but deploy REGION is '$REGION'${NC}"
    echo "Make sure this is intentional. INSTANCE_CONNECTION_NAME: $INSTANCE_CONNECTION_NAME"
fi

if [ "$INSTANCE_STATE" != "RUNNABLE" ]; then
    echo -e "${RED}❌ Cloud SQL instance is not RUNNABLE (state: $INSTANCE_STATE)${NC}"
    echo "Please ensure the database is ready before deploying."
    exit 1
fi

echo -e "${GREEN}✅ Cloud SQL instance is RUNNABLE${NC}"
echo ""

# Step 4: Check Secret Manager
echo -e "${YELLOW}📋 Step 4: Checking Secret Manager...${NC}"

# DATABASE_URL
if ! gcloud secrets describe zeno-auth-database-url >/dev/null 2>&1; then
    echo -e "${RED}❌ Secret 'zeno-auth-database-url' not found${NC}"
    echo "Example to create:"
    echo "  echo -n 'postgres://$DB_USER:***@/zeno_auth?host=/cloudsql/$INSTANCE_CONNECTION_NAME&sslmode=disable' \\"
    echo "    | gcloud secrets create zeno-auth-database-url --data-file=-"
    exit 1
fi

DB_SECRET_VERSION=$(gcloud secrets versions list zeno-auth-database-url \
  --format="value(name)" --limit=1)
if [ -z "$DB_SECRET_VERSION" ]; then
    echo -e "${RED}❌ No DATABASE_URL secret versions found${NC}"
    exit 1
fi
echo -e "${GREEN}✅ DATABASE_URL secret exists (version: $DB_SECRET_VERSION)${NC}"

# JWT_PRIVATE_KEY
if ! gcloud secrets describe zeno-auth-jwt-private-key >/dev/null 2>&1; then
    echo -e "${RED}❌ Secret 'zeno-auth-jwt-private-key' not found${NC}"
    echo "Create RSA key:"
    echo "  openssl genrsa 2048 | gcloud secrets create zeno-auth-jwt-private-key --data-file=-"
    exit 1
fi

JWT_SECRET_VERSION=$(gcloud secrets versions list zeno-auth-jwt-private-key \
  --format="value(name)" --limit=1 2>/dev/null)
if [ -z "$JWT_SECRET_VERSION" ]; then
    echo -e "${RED}❌ No JWT_PRIVATE_KEY secret versions found${NC}"
    exit 1
fi
echo -e "${GREEN}✅ JWT_PRIVATE_KEY secret exists (version: $JWT_SECRET_VERSION)${NC}"

# SENDGRID_API_KEY (optional but used in deploy)
if ! gcloud secrets describe zeno-auth-sendgrid-api-key >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Secret 'zeno-auth-sendgrid-api-key' not found${NC}"
    echo "If you plan to send emails, create it:"
    echo "  echo -n 'SG.***' | gcloud secrets create zeno-auth-sendgrid-api-key --data-file=-"
fi

echo ""

# Step 5: Check Artifact Registry repository
echo -e "${YELLOW}📋 Step 5: Checking Artifact Registry...${NC}"
if ! gcloud artifacts repositories describe "$REPO_NAME" --location="$REGION" >/dev/null 2>&1; then
    echo "Creating Artifact Registry repository '$REPO_NAME' in $REGION..."
    gcloud artifacts repositories create "$REPO_NAME" \
        --repository-format=docker \
        --location="$REGION" \
        --description="Zeno Auth Docker images"
    echo -e "${GREEN}✅ Repository created${NC}"
else
    echo -e "${GREEN}✅ Repository '$REPO_NAME' exists${NC}"
fi
echo ""

# Step 6: Build and push Docker image
echo -e "${YELLOW}📋 Step 6: Building and pushing Docker image...${NC}"
IMAGE_LATEST="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO_NAME/zeno-auth:latest"

echo "Building image: $IMAGE_LATEST"
gcloud builds submit --tag "$IMAGE_LATEST"

echo -e "${GREEN}✅ Image built and pushed${NC}"
echo ""

# Common deploy flags
COMMON_DEPLOY_FLAGS=(
  --image="$IMAGE_LATEST"
  --region="$REGION"
  --platform=managed
  --service-account="$SERVICE_ACCOUNT"
  --add-cloudsql-instances="$INSTANCE_CONNECTION_NAME"
  --set-secrets=DATABASE_URL=zeno-auth-database-url:latest,JWT_PRIVATE_KEY=zeno-auth-jwt-private-key:latest,SENDGRID_API_KEY=zeno-auth-sendgrid-api-key:latest
  --set-env-vars=ENV="$APP_ENV",APP_NAME=zeno-auth,CORS_ALLOWED_ORIGINS=https://storage.googleapis.com,EMAIL_FROM=noreply@em2292.zeno-cy.com,EMAIL_FROM_NAME=ZenoN-Cloud,APP_BASE_URL="$APP_BASE_URL",FRONTEND_BASE_URL="$APP_BASE_URL/index.html",LOG_FILE=/tmp/zeno-auth.log
  --port=8080
  --memory=512Mi
  --cpu=1
  --timeout=300
  --max-instances=10
  --min-instances=0
  --concurrency=80
  --allow-unauthenticated
)

# Step 7: Deploy to Cloud Run
echo -e "${YELLOW}📋 Step 7: Deploying to Cloud Run...${NC}"

if gcloud run services describe "$SERVICE_NAME" --region="$REGION" >/dev/null 2>&1; then
    echo "Updating existing service '$SERVICE_NAME'..."
    gcloud run deploy "$SERVICE_NAME" "${COMMON_DEPLOY_FLAGS[@]}"
else
    echo "Creating new service '$SERVICE_NAME'..."
    gcloud run deploy "$SERVICE_NAME" "${COMMON_DEPLOY_FLAGS[@]}"
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

if curl -s -f "$SERVICE_URL/health" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    # Если health возвращает JSON
    if command -v jq >/dev/null 2>&1; then
        curl -s "$SERVICE_URL/health" | jq .
    else
        curl -s "$SERVICE_URL/health"
        echo
    fi
else
    echo -e "${RED}❌ Health check failed${NC}"
    echo "Посмотри логи, например:"
    echo "  gcloud logging read \"resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME\" \\"
    echo "    --region=$REGION --limit=50 --format=\"value(timestamp, textPayload, jsonPayload)\""
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}🎉 Deployment Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Service URL: $SERVICE_URL"
echo "Image:       $IMAGE_LATEST"
echo ""
echo "Next steps:"
echo "  1. Test API:  curl $SERVICE_URL/health"
echo "  2. View logs: gcloud logging read \"resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME\" --limit=50"
echo "  3. Console:   https://console.cloud.google.com/run/detail/$REGION/$SERVICE_NAME"
echo ""
