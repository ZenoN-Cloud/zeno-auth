# 🚀 GCP Cloud Run Deployment Guide

**Version:** 1.1.0  
**Target:** Google Cloud Platform (Cloud Run + Cloud SQL)

---

## ✅ Pre-Deployment Checklist

### 1. Environment Variables

Ensure these are set before deployment:

```bash
export PROJECT_ID=zeno-cy-dev-001
export REGION=europe-west3
export INSTANCE_ID=zeno-auth-db-dev
export DB_NAME=zeno_auth
export DB_USER=zeno_auth_app
export DB_PASSWORD=<from-1password>
export INSTANCE_CONNECTION_NAME=zeno-cy-dev-001:europe-west3:zeno-auth-db-dev
```

### 2. Cloud SQL Status

Check database is ready:

```bash
gcloud sql instances describe $INSTANCE_ID --format="value(state)"
```

**Expected:** `RUNNABLE`

### 3. Secret Manager

Verify secret exists:

```bash
gcloud secrets describe zeno-auth-database-url
```

**Expected:**
```
name: projects/zeno-cy-dev-001/secrets/zeno-auth-database-url
replication: automatic
```

Check version:

```bash
gcloud secrets versions list zeno-auth-database-url
```

**Expected:** At least version `1` with state `enabled`

### 4. 1Password Vault

**Vault:** ZenoN  
**Entry:** Zeno Auth Database (dev)

**Required fields:**
- username: `zeno_auth_app`
- password: `<secure-password>`
- database: `zeno_auth`
- hostname: `zeno-auth-db-dev`
- region: `europe-west3`
- project: `zeno-cy-dev-001`
- connection_name: `zeno-cy-dev-001:europe-west3:zeno-auth-db-dev`
- database_url: `postgres://zeno_auth_app:***@/zeno_auth?host=/cloudsql/...`

---

## 🚀 Deployment Methods

### Method 1: Automated Script (Recommended)

```bash
cd deploy
./gcp-deploy.sh
```

This script will:
1. ✅ Check prerequisites
2. ✅ Verify Cloud SQL is ready
3. ✅ Check Secret Manager
4. ✅ Create Artifact Registry (if needed)
5. ✅ Build and push Docker image
6. ✅ Deploy to Cloud Run
7. ✅ Run health check

### Method 2: Manual Deployment

#### Step 1: Create Artifact Registry

```bash
gcloud artifacts repositories create zeno-auth \
  --repository-format=docker \
  --location="$REGION"
```

#### Step 2: Build and Push Image

```bash
export IMAGE="europe-west3-docker.pkg.dev/$PROJECT_ID/zeno-auth/zeno-auth:dev-$(date +%Y%m%d)"

gcloud builds submit --tag "$IMAGE"
```

#### Step 3: Deploy to Cloud Run

**First deployment (creates service):**

```bash
gcloud run deploy zeno-auth-dev \
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
```

**Update existing service:**

```bash
gcloud run services update zeno-auth-dev \
  --image="$IMAGE" \
  --region="$REGION"
```

---

## 🔧 Configuration Updates

### Update Cloud SQL Connection

```bash
gcloud run services update zeno-auth-dev \
  --add-cloudsql-instances="$INSTANCE_CONNECTION_NAME" \
  --region="$REGION"
```

### Update Secret

```bash
gcloud run services update zeno-auth-dev \
  --set-secrets=DATABASE_URL=zeno-auth-database-url:latest \
  --region="$REGION"
```

### Update Environment Variables

```bash
gcloud run services update zeno-auth-dev \
  --set-env-vars="ENV=production,LOG_LEVEL=info" \
  --region="$REGION"
```

---

## 🧪 Verification

### 1. Check Service Status

```bash
gcloud run services describe zeno-auth-dev --region="$REGION"
```

### 2. View Logs

```bash
gcloud logs read zeno-auth-dev --region="$REGION" --limit=50
```

**Successful connection log:**
```
Connected to PostgreSQL via unix socket /cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev
```

### 3. Health Check

Get service URL:

```bash
SERVICE_URL=$(gcloud run services describe zeno-auth-dev --region="$REGION" --format="value(status.url)")
```

Test endpoints:

```bash
# Basic health
curl $SERVICE_URL/health

# Readiness check
curl $SERVICE_URL/health/ready

# Metrics
curl $SERVICE_URL/metrics
```

**Expected response:**
```json
{
  "service": "zeno-auth",
  "status": "healthy"
}
```

---

## 🔐 Security Configuration

### 1. Service Account

Create dedicated service account:

```bash
gcloud iam service-accounts create zeno-auth-sa \
  --display-name="Zeno Auth Service Account"
```

Grant Cloud SQL Client role:

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

Grant Secret Manager accessor:

```bash
gcloud secrets add-iam-policy-binding zeno-auth-database-url \
  --member="serviceAccount:zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

Update Cloud Run service:

```bash
gcloud run services update zeno-auth-dev \
  --service-account="zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --region="$REGION"
```

### 2. IAM Policies

Restrict access to authenticated users only:

```bash
gcloud run services update zeno-auth-dev \
  --no-allow-unauthenticated \
  --region="$REGION"
```

---

## 📊 Monitoring

### Cloud Console

**Service Dashboard:**
```
https://console.cloud.google.com/run/detail/europe-west3/zeno-auth-dev
```

**Logs Explorer:**
```
https://console.cloud.google.com/logs/query
```

**Metrics:**
```
https://console.cloud.google.com/monitoring
```

### CLI Monitoring

**Real-time logs:**
```bash
gcloud logs tail zeno-auth-dev --region="$REGION"
```

**Error logs:**
```bash
gcloud logs read zeno-auth-dev \
  --region="$REGION" \
  --filter="severity>=ERROR" \
  --limit=50
```

**Request metrics:**
```bash
gcloud monitoring time-series list \
  --filter='metric.type="run.googleapis.com/request_count"'
```

---

## 🔄 Rollback

### Rollback to Previous Revision

List revisions:

```bash
gcloud run revisions list --service=zeno-auth-dev --region="$REGION"
```

Rollback:

```bash
gcloud run services update-traffic zeno-auth-dev \
  --to-revisions=zeno-auth-dev-00001-abc=100 \
  --region="$REGION"
```

---

## 🐛 Troubleshooting

### Issue: Service won't start

**Check logs:**
```bash
gcloud logs read zeno-auth-dev --region="$REGION" --limit=100
```

**Common causes:**
- Database connection string incorrect
- Secret not accessible
- Cloud SQL instance not ready
- Insufficient IAM permissions

### Issue: Database connection failed

**Verify Cloud SQL connector:**
```bash
gcloud run services describe zeno-auth-dev \
  --region="$REGION" \
  --format="value(spec.template.spec.containers[0].env)"
```

**Check secret value:**
```bash
gcloud secrets versions access latest --secret="zeno-auth-database-url"
```

### Issue: 502 Bad Gateway

**Possible causes:**
- Service crashed on startup
- Health check endpoint not responding
- Port mismatch (ensure --port=8080)

**Solution:**
```bash
# Check container logs
gcloud logs read zeno-auth-dev --region="$REGION" --limit=50

# Verify port configuration
gcloud run services describe zeno-auth-dev \
  --region="$REGION" \
  --format="value(spec.template.spec.containers[0].ports[0].containerPort)"
```

---

## 📝 Makefile Integration

Add to `Makefile`:

```makefile
# GCP Deployment
gcp-deploy: ## Deploy to GCP Cloud Run
	@./deploy/gcp-deploy.sh

gcp-logs: ## View GCP logs
	@gcloud logs tail zeno-auth-dev --region=europe-west3

gcp-status: ## Check GCP service status
	@gcloud run services describe zeno-auth-dev --region=europe-west3
```

Usage:

```bash
make gcp-deploy
make gcp-logs
make gcp-status
```

---

## ✅ Post-Deployment Checklist

- [ ] Service is running (status: READY)
- [ ] Health check passes
- [ ] Database connection works
- [ ] Logs show no errors
- [ ] Metrics are being collected
- [ ] IAM permissions configured
- [ ] Service account assigned
- [ ] Secrets accessible
- [ ] CORS configured (if needed)
- [ ] Rate limiting active
- [ ] Monitoring alerts set up

---

## 🔗 Useful Links

- **Cloud Run Console:** https://console.cloud.google.com/run
- **Cloud SQL Console:** https://console.cloud.google.com/sql
- **Secret Manager:** https://console.cloud.google.com/security/secret-manager
- **Logs Explorer:** https://console.cloud.google.com/logs
- **Monitoring:** https://console.cloud.google.com/monitoring

---

## 📞 Support

For deployment issues:
1. Check logs: `gcloud logs read zeno-auth-dev --region=europe-west3 --limit=100`
2. Verify checklist above
3. Review error messages
4. Check GCP status: https://status.cloud.google.com/

---

**Status:** 🟢 READY FOR DEPLOYMENT  
**Last Updated:** 2024-11-22  
**Version:** 1.1.0
