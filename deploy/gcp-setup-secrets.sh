#!/bin/bash
# GCP Secrets & IAM Setup Script for Zeno Auth
# Run this BEFORE first deployment

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
SERVICE_ACCOUNT="${SERVICE_ACCOUNT:-zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🔐 Zeno Auth - GCP Secrets & IAM Setup${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo "Service Account: $SERVICE_ACCOUNT"
echo ""

# Step 1: Set project
echo -e "${YELLOW}📋 Step 1: Setting GCP project...${NC}"
gcloud config set project "$PROJECT_ID"
echo -e "${GREEN}✅ Project set${NC}"
echo ""

# Step 2: Create Service Account (if not exists)
echo -e "${YELLOW}📋 Step 2: Checking Service Account...${NC}"
if gcloud iam service-accounts describe "$SERVICE_ACCOUNT" &> /dev/null; then
    echo -e "${GREEN}✅ Service Account exists${NC}"
else
    echo "Creating Service Account..."
    gcloud iam service-accounts create zeno-auth-sa \
        --display-name="Zeno Auth Service Account" \
        --description="Service account for Zeno Auth Cloud Run service"
    echo -e "${GREEN}✅ Service Account created${NC}"
fi
echo ""

# Step 3: Grant IAM Roles
echo -e "${YELLOW}📋 Step 3: Granting IAM roles...${NC}"

# Cloud SQL Client
echo "Granting roles/cloudsql.client..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/cloudsql.client" \
    --condition=None \
    > /dev/null 2>&1 || true

# Secret Manager Accessor
echo "Granting roles/secretmanager.secretAccessor..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/secretmanager.secretAccessor" \
    --condition=None \
    > /dev/null 2>&1 || true

# Logging Writer
echo "Granting roles/logging.logWriter..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/logging.logWriter" \
    --condition=None \
    > /dev/null 2>&1 || true

# Monitoring Metric Writer
echo "Granting roles/monitoring.metricWriter..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/monitoring.metricWriter" \
    --condition=None \
    > /dev/null 2>&1 || true

echo -e "${GREEN}✅ IAM roles granted${NC}"
echo ""

# Step 4: Verify IAM roles
echo -e "${YELLOW}📋 Step 4: Verifying IAM roles...${NC}"
ROLES=$(gcloud projects get-iam-policy "$PROJECT_ID" \
    --flatten="bindings[].members" \
    --filter="bindings.members:serviceAccount:$SERVICE_ACCOUNT" \
    --format="table(bindings.role)")

echo "$ROLES"
echo ""

# Step 5: Setup DATABASE_URL Secret
echo -e "${YELLOW}📋 Step 5: Setting up DATABASE_URL secret...${NC}"

if gcloud secrets describe zeno-auth-database-url &> /dev/null; then
    echo -e "${GREEN}✅ DATABASE_URL secret already exists${NC}"
    echo "Current version:"
    gcloud secrets versions list zeno-auth-database-url --limit=1
else
    echo -e "${YELLOW}⚠️  DATABASE_URL secret not found${NC}"
    echo ""
    echo "Please provide the DATABASE_URL connection string:"
    echo "Format: postgres://USER:PASSWORD@/DB_NAME?host=/cloudsql/INSTANCE_CONNECTION_NAME&sslmode=disable"
    echo ""
    echo "Example:"
    echo "postgres://zeno_auth_app:MyPassword123@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev&sslmode=disable"
    echo ""
    read -p "Enter DATABASE_URL (or press Enter to skip): " DATABASE_URL
    
    if [ -n "$DATABASE_URL" ]; then
        echo -n "$DATABASE_URL" | gcloud secrets create zeno-auth-database-url \
            --data-file=- \
            --replication-policy="automatic"
        echo -e "${GREEN}✅ DATABASE_URL secret created${NC}"
    else
        echo -e "${YELLOW}⚠️  Skipped. You'll need to create it manually before deployment.${NC}"
    fi
fi
echo ""

# Step 6: Setup JWT_PRIVATE_KEY Secret
echo -e "${YELLOW}📋 Step 6: Setting up JWT_PRIVATE_KEY secret...${NC}"

if gcloud secrets describe zeno-auth-jwt-private-key &> /dev/null; then
    echo -e "${GREEN}✅ JWT_PRIVATE_KEY secret already exists${NC}"
    echo "Current version:"
    gcloud secrets versions list zeno-auth-jwt-private-key --limit=1
else
    echo -e "${YELLOW}⚠️  JWT_PRIVATE_KEY secret not found${NC}"
    echo ""
    read -p "Generate new JWT keys? (y/n): " GENERATE_JWT
    
    if [ "$GENERATE_JWT" = "y" ] || [ "$GENERATE_JWT" = "Y" ]; then
        echo "Generating RSA key pair..."
        
        # Generate private key
        TEMP_PRIVATE=$(mktemp)
        TEMP_PUBLIC=$(mktemp)
        
        openssl genrsa -out "$TEMP_PRIVATE" 2048
        openssl rsa -in "$TEMP_PRIVATE" -pubout -out "$TEMP_PUBLIC"
        
        # Create secrets
        gcloud secrets create zeno-auth-jwt-private-key \
            --data-file="$TEMP_PRIVATE" \
            --replication-policy="automatic"
        
        gcloud secrets create zeno-auth-jwt-public-key \
            --data-file="$TEMP_PUBLIC" \
            --replication-policy="automatic"
        
        echo -e "${GREEN}✅ JWT keys generated and stored${NC}"
        echo ""
        echo "Public key (save this for verification):"
        cat "$TEMP_PUBLIC"
        
        # Cleanup
        rm -f "$TEMP_PRIVATE" "$TEMP_PUBLIC"
    else
        echo -e "${YELLOW}⚠️  Skipped. You'll need to create it manually before deployment.${NC}"
        echo "To create manually:"
        echo "  openssl genrsa 2048 | gcloud secrets create zeno-auth-jwt-private-key --data-file=-"
    fi
fi
echo ""

# Step 7: Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}🎉 Setup Complete!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Summary:"
echo "  ✅ Service Account: $SERVICE_ACCOUNT"
echo "  ✅ IAM Roles: cloudsql.client, secretmanager.secretAccessor, logging.logWriter, monitoring.metricWriter"

# Check secrets
if gcloud secrets describe zeno-auth-database-url &> /dev/null; then
    echo "  ✅ DATABASE_URL secret: exists"
else
    echo "  ⚠️  DATABASE_URL secret: NOT CREATED"
fi

if gcloud secrets describe zeno-auth-jwt-private-key &> /dev/null; then
    echo "  ✅ JWT_PRIVATE_KEY secret: exists"
else
    echo "  ⚠️  JWT_PRIVATE_KEY secret: NOT CREATED"
fi

echo ""
echo "Next steps:"
echo "  1. Verify Cloud SQL instance is running"
echo "  2. Run: ./deploy/gcp-deploy.sh"
echo "  3. Test: curl \$SERVICE_URL/health"
echo ""
