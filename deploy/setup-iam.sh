#!/bin/bash
# Quick IAM Setup for Zeno Auth

set -e

PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
SERVICE_ACCOUNT="zeno-auth-sa@${PROJECT_ID}.iam.gserviceaccount.com"

echo "🔐 Setting up IAM for Zeno Auth"
echo "Project: $PROJECT_ID"
echo "Service Account: $SERVICE_ACCOUNT"
echo ""

# Create Service Account
echo "Creating Service Account..."
if gcloud iam service-accounts describe "$SERVICE_ACCOUNT" &> /dev/null; then
    echo "✅ Service Account already exists"
else
    gcloud iam service-accounts create zeno-auth-sa \
        --display-name="Zeno Auth Service Account" \
        --description="Service account for Zeno Auth Cloud Run"
    echo "✅ Service Account created"
fi

# Grant roles
echo ""
echo "Granting IAM roles..."

echo "  - Cloud SQL Client..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/cloudsql.client" \
    --condition=None > /dev/null

echo "  - Secret Manager Accessor..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/secretmanager.secretAccessor" \
    --condition=None > /dev/null

echo "  - Logging Writer..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/logging.logWriter" \
    --condition=None > /dev/null

echo "  - Monitoring Metric Writer..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/monitoring.metricWriter" \
    --condition=None > /dev/null

echo ""
echo "✅ All IAM roles granted"
echo ""
echo "Verification:"
gcloud projects get-iam-policy "$PROJECT_ID" \
    --flatten="bindings[].members" \
    --filter="bindings.members:serviceAccount:$SERVICE_ACCOUNT" \
    --format="table(bindings.role)"

echo ""
echo "✅ IAM setup complete!"
