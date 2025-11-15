# Deployment Guide

## Prerequisites

1. GCP projects created: `zenon-cloud-dev-001`, `zenon-cloud-prod-001`
2. GitHub repository with Workload Identity Federation configured
3. Required secrets in GitHub:
   - `WIF_PROVIDER` - Workload Identity Provider for dev
   - `WIF_SERVICE_ACCOUNT` - Service account for dev
   - `WIF_PROVIDER_PROD` - Workload Identity Provider for prod  
   - `WIF_SERVICE_ACCOUNT_PROD` - Service account for prod

## Setup Steps

1. **Run GCP setup:**
   ```bash
   # Follow instructions in gcp-setup.md
   ./deploy/gcp-setup.md
   ```

2. **Configure GitHub Environments:**
   - Create `dev` environment (auto-deploy)
   - Create `prod` environment (manual approval required)

3. **Deploy to dev:**
   ```bash
   git push origin main
   ```

4. **Deploy to prod:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

## Manual Deployment

```bash
# Build and deploy manually
gcloud config set project zenon-cloud-dev-001
gcloud builds submit --tag europe-west3-docker.pkg.dev/zenon-cloud-dev-001/zeno-auth-repo/zeno-auth:manual

gcloud run deploy zeno-auth-dev \
  --image=europe-west3-docker.pkg.dev/zenon-cloud-dev-001/zeno-auth-repo/zeno-auth:manual \
  --region=europe-west3
```

## Run Migrations

```bash
# Get database URL from secret
DATABASE_URL=$(gcloud secrets versions access latest --secret="zeno-auth-db-dsn")

# Run migrations
./deploy/migrate.sh up "$DATABASE_URL"
```

## Monitoring

- **Health check:** `https://zeno-auth-dev-xxx.run.app/health`
- **Logs:** `gcloud logs read --service=zeno-auth-dev`
- **Metrics:** Cloud Run console