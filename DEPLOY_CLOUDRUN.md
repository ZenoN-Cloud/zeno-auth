# Deploy zeno-auth to Cloud Run

## Prerequisites

1. **GCP Project Setup**
   ```bash
   export GCP_PROJECT_ID="your-project-id"
   export GCP_REGION="europe-west1"
   ```

2. **Enable Required APIs**
   ```bash
   gcloud services enable run.googleapis.com \
     cloudbuild.googleapis.com \
     secretmanager.googleapis.com \
     sqladmin.googleapis.com
   ```

3. **Create Cloud SQL Instance** (if not exists)
   ```bash
   gcloud sql instances create zeno-auth-db \
     --database-version=POSTGRES_15 \
     --tier=db-f1-micro \
     --region=${GCP_REGION} \
     --project=${GCP_PROJECT_ID}
   ```

4. **Create Database**
   ```bash
   gcloud sql databases create zeno_auth \
     --instance=zeno-auth-db \
     --project=${GCP_PROJECT_ID}
   ```

5. **Create Secrets**
   
   **Database URL:**
   ```bash
   echo -n "postgres://user:password@/zeno_auth?host=/cloudsql/${GCP_PROJECT_ID}:${GCP_REGION}:zeno-auth-db" | \
     gcloud secrets create zeno-auth-database-url \
       --data-file=- \
       --project=${GCP_PROJECT_ID}
   ```

   **JWT Private Key:**
   ```bash
   cat keys/private.pem | \
     gcloud secrets create zeno-auth-jwt-private-key \
       --data-file=- \
       --project=${GCP_PROJECT_ID}
   ```

## Deploy

```bash
./deploy-cloudrun.sh
```

## Environment Variables

Required secrets (stored in Secret Manager):
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_PRIVATE_KEY` - RSA private key for JWT signing

Optional environment variables:
- `ENV` - Environment (production/staging/dev)
- `PORT` - HTTP port (default: 8080)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)
- `ACCESS_TOKEN_TTL` - Access token TTL in seconds (default: 1800)
- `REFRESH_TOKEN_TTL` - Refresh token TTL in seconds (default: 1209600)
- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed origins

## Verify Deployment

```bash
SERVICE_URL=$(gcloud run services describe zeno-auth \
  --platform managed \
  --region ${GCP_REGION} \
  --project ${GCP_PROJECT_ID} \
  --format 'value(status.url)')

curl ${SERVICE_URL}/health
```

## Logs

```bash
gcloud run logs read zeno-auth \
  --region ${GCP_REGION} \
  --project ${GCP_PROJECT_ID} \
  --limit 50
```
