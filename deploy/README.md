# 🚀 Deployment Guide

Production deployment guide for Zeno Auth on GCP Cloud Run.

## Prerequisites

- GCP account with billing enabled
- `gcloud` CLI installed and configured
- Docker installed (for local testing)
- Project ID: `zeno-cy-dev-001`
- Region: `europe-west3`

## Quick Deploy

```bash
# 1. Setup infrastructure (one-time)
./deploy/gcp-setup-complete.sh

# 2. Deploy application
./deploy/gcp-deploy.sh

# 3. Check status
./deploy/gcp-status-check.sh
```

## Infrastructure Setup

The `gcp-setup-complete.sh` script creates:

1. **Cloud SQL PostgreSQL 17**
    - Instance: `zeno-auth-db-dev`
    - Database: `zeno_auth`
    - User: `zeno_auth_app`

2. **Secret Manager**
    - `zeno-auth-database-url` - Connection string
    - `zeno-auth-jwt-private-key` - RSA private key

3. **Service Account**
    - Email: `zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com`
    - Roles: Cloud SQL Client, Secret Accessor, Log Writer, Metric Writer

4. **Artifact Registry**
    - Repository: `europe-west3-docker.pkg.dev/zeno-cy-dev-001/zeno-auth`

## Manual Setup

### 1. Enable APIs

```bash
gcloud services enable \
  sqladmin.googleapis.com \
  run.googleapis.com \
  secretmanager.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com
```

### 2. Create Cloud SQL

```bash
gcloud sql instances create zeno-auth-db-dev \
  --database-version=POSTGRES_17 \
  --tier=db-f1-micro \
  --region=europe-west3 \
  --storage-type=SSD \
  --storage-size=10GB

gcloud sql databases create zeno_auth \
  --instance=zeno-auth-db-dev

gcloud sql users create zeno_auth_app \
  --instance=zeno-auth-db-dev \
  --password=<STRONG_PASSWORD>
```

### 3. Create Secrets

```bash
# Database URL
echo -n "postgres://user:pass@/zeno_auth?host=/cloudsql/PROJECT:REGION:INSTANCE" | \
  gcloud secrets create zeno-auth-database-url --data-file=-

# JWT Keys
openssl genrsa 2048 | \
  gcloud secrets create zeno-auth-jwt-private-key --data-file=-
```

### 4. Create Service Account

```bash
gcloud iam service-accounts create zeno-auth-sa \
  --display-name="Zeno Auth Service Account"

# Grant roles
for role in cloudsql.client secretmanager.secretAccessor logging.logWriter monitoring.metricWriter; do
  gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
    --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
    --role="roles/$role"
done
```

### 5. Deploy to Cloud Run

```bash
# Build and push image
gcloud builds submit --tag europe-west3-docker.pkg.dev/zeno-cy-dev-001/zeno-auth/zeno-auth:latest

# Deploy service
gcloud run deploy zeno-auth-dev \
  --image=europe-west3-docker.pkg.dev/zeno-cy-dev-001/zeno-auth/zeno-auth:latest \
  --region=europe-west3 \
  --platform=managed \
  --service-account=zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com \
  --add-cloudsql-instances=zeno-cy-dev-001:europe-west3:zeno-auth-db-dev \
  --set-secrets=DATABASE_URL=zeno-auth-database-url:latest,JWT_PRIVATE_KEY=zeno-auth-jwt-private-key:latest \
  --set-env-vars=ENV=production,APP_NAME=zeno-auth,PORT=8080 \
  --port=8080 \
  --memory=512Mi \
  --cpu=1 \
  --timeout=300 \
  --max-instances=10 \
  --min-instances=0 \
  --concurrency=80 \
  --allow-unauthenticated
```

## Configuration

### Environment Variables

| Variable               | Description                          | Required |
|------------------------|--------------------------------------|----------|
| `DATABASE_URL`         | PostgreSQL connection string         | Yes      |
| `JWT_PRIVATE_KEY`      | RSA private key (PEM format)         | Yes      |
| `ENV`                  | Environment (production/development) | Yes      |
| `PORT`                 | HTTP port                            | Yes      |
| `APP_NAME`             | Application name                     | No       |
| `CORS_ALLOWED_ORIGINS` | Comma-separated origins              | No       |

### Cloud Run Settings

- **Memory:** 512Mi (minimum for Go app)
- **CPU:** 1 (sufficient for auth service)
- **Timeout:** 300s (5 minutes)
- **Max Instances:** 10 (adjust based on load)
- **Min Instances:** 0 (cold start acceptable)
- **Concurrency:** 80 (requests per instance)

## Monitoring

### Health Checks

```bash
SERVICE_URL=$(gcloud run services describe zeno-auth-dev --region=europe-west3 --format="value(status.url)")

# Basic health
curl $SERVICE_URL/health

# Readiness (includes DB check)
curl $SERVICE_URL/health/ready

# Liveness
curl $SERVICE_URL/health/live
```

### Logs

```bash
# Tail logs
gcloud logs tail zeno-auth-dev --region=europe-west3

# Filter errors
gcloud logs read zeno-auth-dev --region=europe-west3 --filter="severity>=ERROR" --limit=50

# View specific request
gcloud logs read zeno-auth-dev --region=europe-west3 --filter="labels.request_id=<ID>"
```

### Metrics

```bash
# View metrics
curl $SERVICE_URL/metrics

# In GCP Console
# Cloud Run > zeno-auth-dev > Metrics
```

## Troubleshooting

### Database Connection Issues

```bash
# Check Cloud SQL status
gcloud sql instances describe zeno-auth-db-dev

# Test connection via proxy
cloud-sql-proxy zeno-cy-dev-001:europe-west3:zeno-auth-db-dev &
psql "host=127.0.0.1 port=5432 dbname=zeno_auth user=zeno_auth_app"
```

### Secret Access Issues

```bash
# Verify secret exists
gcloud secrets describe zeno-auth-database-url

# Check service account permissions
gcloud projects get-iam-policy zeno-cy-dev-001 \
  --flatten="bindings[].members" \
  --filter="bindings.members:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com"
```

### Migration Failures

Migrations run automatically on container startup. If they fail:

```bash
# View logs
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=100

# Run migrations manually via Cloud SQL proxy
migrate -path ./migrations -database "$DATABASE_URL" up
```

## Rollback

```bash
# List revisions
gcloud run revisions list --service=zeno-auth-dev --region=europe-west3

# Rollback to previous revision
gcloud run services update-traffic zeno-auth-dev \
  --region=europe-west3 \
  --to-revisions=<REVISION_NAME>=100
```

## Security Checklist

- [ ] Database password is strong (32+ characters)
- [ ] JWT private key is securely generated
- [ ] Secrets are stored in Secret Manager (not in code)
- [ ] Service account has minimal required permissions
- [ ] Cloud SQL has no public IP (Cloud Run connects via Unix socket)
- [ ] CORS origins are properly configured
- [ ] Rate limiting is enabled
- [ ] Audit logging is enabled

## Production Considerations

### Scaling

- Monitor request latency and adjust `--max-instances`
- Consider `--min-instances=1` for warm start
- Adjust `--concurrency` based on request duration

### Database

- Upgrade from `db-f1-micro` for production load
- Enable automated backups
- Set up read replicas for high traffic
- Monitor connection pool usage

### Monitoring

- Set up alerting for errors and latency
- Configure uptime checks
- Monitor Cloud SQL metrics
- Track JWT token usage

### Cost Optimization

- Use `--min-instances=0` for dev/staging
- Set appropriate `--max-instances` limit
- Monitor Cloud Run and Cloud SQL costs
- Consider committed use discounts

## Support

For issues:

1. Check logs: `make gcp-logs`
2. Verify status: `make gcp-status-check`
3. Review [Architecture](../docs/architecture.md)
4. Open GitHub issue

---

**Last Updated:** 2024  
**Version:** 1.1.0
