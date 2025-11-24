#!/bin/bash
# Quick Dev Deployment Script for Zeno Auth
set -e

PROJECT_ID="zeno-cy-dev-001"
REGION="europe-west3"
SERVICE_NAME="zeno-auth-dev"
REPO_NAME="zeno-auth"

echo "🚀 Deploying Zeno Auth to DEV..."
echo "Project: $PROJECT_ID"
echo "Service: $SERVICE_NAME"
echo ""

# Set project
gcloud config set project "$PROJECT_ID"

# Build and deploy
IMAGE="$REGION-docker.pkg.dev/$PROJECT_ID/$REPO_NAME/zeno-auth:latest"
echo "Building image..."
gcloud builds submit --tag "$IMAGE"

echo "Deploying to Cloud Run..."
gcloud run deploy "$SERVICE_NAME" \
  --image="$IMAGE" \
  --region="$REGION" \
  --platform=managed \
  --service-account="zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --add-cloudsql-instances="$PROJECT_ID:$REGION:zeno-auth-db-dev" \
  --set-secrets=DATABASE_URL=zeno-auth-database-url:latest,JWT_PRIVATE_KEY=zeno-auth-jwt-private-key:latest,SENDGRID_API_KEY=zeno-auth-sendgrid-api-key:latest \
  --set-env-vars=ENV=development,APP_NAME=zeno-auth,CORS_ALLOWED_ORIGINS=https://storage.googleapis.com,EMAIL_FROM=noreply@em2292.zeno-cy.com,EMAIL_FROM_NAME=ZenoN-Cloud,APP_BASE_URL=https://storage.googleapis.com/zeno-cy-frontend-dev-001 \
  --port=8080 \
  --memory=512Mi \
  --cpu=1 \
  --max-instances=10 \
  --min-instances=0 \
  --allow-unauthenticated

SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region="$REGION" --format="value(status.url)")
echo ""
echo "✅ Deployed: $SERVICE_URL"
echo "Health: curl $SERVICE_URL/health"
