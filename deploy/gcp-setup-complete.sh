#!/bin/bash
# Complete GCP Setup Script for Zeno Auth
# This script automates the entire GCP infrastructure setup

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
REGION="${REGION:-europe-west3}"
INSTANCE_ID="${INSTANCE_ID:-zeno-auth-db-dev}"
DB_NAME="${DB_NAME:-zeno_auth}"
DB_USER="${DB_USER:-zeno_auth_app}"
SERVICE_ACCOUNT_NAME="zeno-auth-sa"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Zeno Auth - Complete GCP Infrastructure Setup        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Project:${NC} $PROJECT_ID"
echo -e "${GREEN}Region:${NC} $REGION"
echo -e "${GREEN}Database:${NC} $INSTANCE_ID"
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for user confirmation
confirm() {
    read -p "$1 [y/N]: " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]]
}

# Step 0: Prerequisites
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 0: Checking Prerequisites${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if ! command_exists gcloud; then
    echo -e "${RED}❌ gcloud CLI not found${NC}"
    exit 1
fi

if ! command_exists openssl; then
    echo -e "${RED}❌ openssl not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites OK${NC}"
echo ""

# Set project
echo -e "${BLUE}Setting GCP project...${NC}"
gcloud config set project "$PROJECT_ID"
echo ""

# Step 1: Enable APIs
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 1: Enabling Required APIs${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

APIS=(
    "sqladmin.googleapis.com"
    "run.googleapis.com"
    "secretmanager.googleapis.com"
    "artifactregistry.googleapis.com"
    "cloudbuild.googleapis.com"
    "logging.googleapis.com"
    "monitoring.googleapis.com"
)

for api in "${APIS[@]}"; do
    echo -e "${BLUE}Enabling $api...${NC}"
    gcloud services enable "$api" --quiet
done

echo -e "${GREEN}✅ APIs enabled${NC}"
echo ""

# Step 2: Create Cloud SQL Instance
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 2: Creating Cloud SQL Instance${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud sql instances describe "$INSTANCE_ID" &>/dev/null; then
    echo -e "${GREEN}✅ Cloud SQL instance already exists${NC}"
else
    echo -e "${BLUE}Creating Cloud SQL instance (this may take 5-10 minutes)...${NC}"
    gcloud sql instances create "$INSTANCE_ID" \
        --database-version=POSTGRES_17 \
        --tier=db-f1-micro \
        --region="$REGION" \
        --storage-type=SSD \
        --storage-size=10GB \
        --backup-start-time=03:00 \
        --maintenance-window-day=SUN \
        --maintenance-window-hour=4 \
        --database-flags=max_connections=100 \
        --quiet
    
    echo -e "${GREEN}✅ Cloud SQL instance created${NC}"
fi

# Wait for instance to be ready
echo -e "${BLUE}Waiting for instance to be ready...${NC}"
while true; do
    STATE=$(gcloud sql instances describe "$INSTANCE_ID" --format="value(state)")
    if [ "$STATE" = "RUNNABLE" ]; then
        break
    fi
    echo -e "${YELLOW}Instance state: $STATE (waiting...)${NC}"
    sleep 5
done
echo -e "${GREEN}✅ Instance is RUNNABLE${NC}"
echo ""

# Step 3: Create Database
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 3: Creating Database${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud sql databases describe "$DB_NAME" --instance="$INSTANCE_ID" &>/dev/null; then
    echo -e "${GREEN}✅ Database already exists${NC}"
else
    gcloud sql databases create "$DB_NAME" --instance="$INSTANCE_ID"
    echo -e "${GREEN}✅ Database created${NC}"
fi
echo ""

# Step 4: Create Database User
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 4: Creating Database User${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud sql users describe "$DB_USER" --instance="$INSTANCE_ID" &>/dev/null; then
    echo -e "${GREEN}✅ Database user already exists${NC}"
    echo -e "${YELLOW}⚠️  If you need to reset the password, delete the user first${NC}"
else
    # Generate strong password
    DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
    
    gcloud sql users create "$DB_USER" \
        --instance="$INSTANCE_ID" \
        --password="$DB_PASSWORD"
    
    echo -e "${GREEN}✅ Database user created${NC}"
    echo -e "${YELLOW}⚠️  IMPORTANT: Save this password securely!${NC}"
    echo -e "${BLUE}Password: ${DB_PASSWORD}${NC}"
    echo ""
    
    # Save to temporary file
    echo "$DB_PASSWORD" > /tmp/zeno-auth-db-password.txt
    echo -e "${YELLOW}Password saved to: /tmp/zeno-auth-db-password.txt${NC}"
    echo ""
    
    if ! confirm "Have you saved the password?"; then
        echo -e "${RED}❌ Please save the password before continuing${NC}"
        exit 1
    fi
fi
echo ""

# Step 5: Create DATABASE_URL Secret
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 5: Creating DATABASE_URL Secret${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud secrets describe zeno-auth-database-url &>/dev/null; then
    echo -e "${GREEN}✅ DATABASE_URL secret already exists${NC}"
else
    # Read password from file if it exists
    if [ -f /tmp/zeno-auth-db-password.txt ]; then
        DB_PASSWORD=$(cat /tmp/zeno-auth-db-password.txt)
    else
        echo -e "${YELLOW}Enter database password:${NC}"
        read -s DB_PASSWORD
        echo ""
    fi
    
    INSTANCE_CONNECTION_NAME="$PROJECT_ID:$REGION:$INSTANCE_ID"
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@/zeno_auth?host=/cloudsql/${INSTANCE_CONNECTION_NAME}&sslmode=disable"
    
    echo -n "$DATABASE_URL" | gcloud secrets create zeno-auth-database-url \
        --data-file=- \
        --replication-policy=automatic
    
    echo -e "${GREEN}✅ DATABASE_URL secret created${NC}"
fi
echo ""

# Step 6: Create JWT Keys
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 6: Creating JWT Keys${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud secrets describe zeno-auth-jwt-private-key &>/dev/null; then
    echo -e "${GREEN}✅ JWT private key secret already exists${NC}"
else
    echo -e "${BLUE}Generating RSA key pair...${NC}"
    
    # Generate keys in temp directory
    TEMP_DIR=$(mktemp -d)
    openssl genrsa -out "$TEMP_DIR/jwt_private.pem" 2048
    openssl rsa -in "$TEMP_DIR/jwt_private.pem" -pubout -out "$TEMP_DIR/jwt_public.pem"
    
    # Create secret
    cat "$TEMP_DIR/jwt_private.pem" | gcloud secrets create zeno-auth-jwt-private-key \
        --data-file=- \
        --replication-policy=automatic
    
    echo -e "${GREEN}✅ JWT private key secret created${NC}"
    echo -e "${YELLOW}Public key saved to: $TEMP_DIR/jwt_public.pem${NC}"
    echo -e "${YELLOW}(You can embed this in your app or store separately)${NC}"
    
    # Cleanup
    rm -f "$TEMP_DIR/jwt_private.pem"
fi
echo ""

# Step 7: Create Service Account
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 7: Creating Service Account${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud iam service-accounts describe "$SERVICE_ACCOUNT_EMAIL" &>/dev/null; then
    echo -e "${GREEN}✅ Service account already exists${NC}"
else
    gcloud iam service-accounts create "$SERVICE_ACCOUNT_NAME" \
        --display-name="Zeno Auth Service Account" \
        --description="Service account for Zeno Auth Cloud Run service"
    
    echo -e "${GREEN}✅ Service account created${NC}"
fi
echo ""

# Step 8: Grant IAM Roles
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 8: Granting IAM Roles${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

ROLES=(
    "roles/cloudsql.client"
    "roles/secretmanager.secretAccessor"
    "roles/logging.logWriter"
    "roles/monitoring.metricWriter"
)

for role in "${ROLES[@]}"; do
    echo -e "${BLUE}Granting $role...${NC}"
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
        --role="$role" \
        --quiet
done

echo -e "${GREEN}✅ IAM roles granted${NC}"
echo ""

# Step 9: Create Artifact Registry
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Step 9: Creating Artifact Registry${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if gcloud artifacts repositories describe zeno-auth --location="$REGION" &>/dev/null; then
    echo -e "${GREEN}✅ Artifact Registry already exists${NC}"
else
    gcloud artifacts repositories create zeno-auth \
        --repository-format=docker \
        --location="$REGION" \
        --description="Zeno Auth Docker images"
    
    echo -e "${GREEN}✅ Artifact Registry created${NC}"
fi
echo ""

# Cleanup temporary files
if [ -f /tmp/zeno-auth-db-password.txt ]; then
    rm -f /tmp/zeno-auth-db-password.txt
fi

# Summary
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ GCP Infrastructure Setup Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}Summary:${NC}"
echo -e "  • Cloud SQL Instance: ${GREEN}$INSTANCE_ID${NC}"
echo -e "  • Database: ${GREEN}$DB_NAME${NC}"
echo -e "  • Database User: ${GREEN}$DB_USER${NC}"
echo -e "  • Service Account: ${GREEN}$SERVICE_ACCOUNT_EMAIL${NC}"
echo -e "  • Artifact Registry: ${GREEN}$REGION-docker.pkg.dev/$PROJECT_ID/zeno-auth${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo -e "  1. Deploy the application:"
echo -e "     ${BLUE}./deploy/gcp-deploy.sh${NC}"
echo ""
echo -e "  2. Test the deployment:"
echo -e "     ${BLUE}make gcp-health${NC}"
echo ""
echo -e "  3. View logs:"
echo -e "     ${BLUE}make gcp-logs${NC}"
echo ""
