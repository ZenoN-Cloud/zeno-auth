# 🚀 Deploy Zeno Auth to GCP Cloud Run

**Status:** ✅ Ready to Deploy  
**Time Required:** 5-10 minutes

---

## ✅ Pre-Deployment Status

All critical blockers (P0) have been fixed:

- ✅ **P0.1** - `migrate` binary in Docker image
- ✅ **P0.2** - `DATABASE_URL` secret created
- ✅ **P0.3** - Cloud SQL connection configured
- ✅ **P0.4** - Service Account + IAM roles granted
- ✅ **P0.5** - `JWT_PRIVATE_KEY` secret created
- ✅ **P0.6** - Debug endpoint secured in production

---

## 🎯 Quick Deploy (3 Commands)

### 1. Verify Cloud SQL Instance

```bash
gcloud sql instances describe zeno-auth-db-dev
```

**Expected:** `state: RUNNABLE`

If not running, start it:
```bash
gcloud sql instances patch zeno-auth-db-dev --activation-policy=ALWAYS
```

---

### 2. (Optional) Run Pre-Deployment Check

```bash
./pre-deploy-check.sh
```

This will verify:
- gcloud CLI installed
- Docker running
- Cloud SQL instance status
- Service Account exists
- IAM roles granted
- Secrets exist
- No secrets in git

---

### 3. Deploy to Cloud Run

```bash
./gcp-deploy.sh
```

**What it does:**
1. ✅ Builds Docker image
2. ✅ Pushes to Artifact Registry
3. ✅ Deploys to Cloud Run with:
   - Service Account
   - Cloud SQL connection
   - DATABASE_URL secret
   - JWT_PRIVATE_KEY secret
   - ENV variables
4. ✅ Runs health check

**Duration:** 5-7 minutes

---

## 🔍 Verify Deployment

After deployment completes:

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe zeno-auth-dev \
  --region=europe-west3 \
  --format="value(status.url)")

echo "Service URL: $SERVICE_URL"

# Test health endpoint
curl $SERVICE_URL/health

# Expected response:
# {"status":"alive","timestamp":"2024-..."}

# Test readiness (includes DB check)
curl $SERVICE_URL/health/ready

# Expected response:
# {"status":"ready","db":"up","timestamp":"2024-..."}
```

---

## 📊 Deployment Configuration

### Project Settings
- **Project ID:** `zeno-cy-dev-001`
- **Region:** `europe-west3`
- **Service Name:** `zeno-auth-dev`
- **Cloud SQL Instance:** `zeno-auth-db-dev`

### Cloud Run Settings
- **Memory:** 512Mi
- **CPU:** 1 vCPU
- **Timeout:** 300s (5 minutes)
- **Max Instances:** 10
- **Min Instances:** 0
- **Concurrency:** 80 requests per instance

### Secrets (from Secret Manager)
- `DATABASE_URL` → `zeno-auth-database-url:latest`
- `JWT_PRIVATE_KEY` → `zeno-auth-jwt-private-key:latest`

### Environment Variables
- `ENV=production`
- `APP_NAME=zeno-auth`
- `PORT=8080`

---

## 🐛 Troubleshooting

### Issue: "Migration failed"

**Check logs:**
```bash
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=50
```

**Common causes:**
1. DATABASE_URL format incorrect
2. Cloud SQL instance not connected
3. Database user permissions missing

**Fix:**
```bash
# Verify DATABASE_URL
gcloud secrets versions access latest --secret=zeno-auth-database-url

# Should be:
# postgres://zeno_auth_app:PASSWORD@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev&sslmode=disable
```

---

### Issue: "Permission denied"

**Check IAM roles:**
```bash
gcloud projects get-iam-policy zeno-cy-dev-001 \
  --flatten="bindings[].members" \
  --filter="bindings.members:zeno-auth-sa"
```

**Fix:**
```bash
./setup-iam.sh
```

---

### Issue: "Health check fails"

**Check service status:**
```bash
gcloud run services describe zeno-auth-dev --region=europe-west3
```

**View real-time logs:**
```bash
gcloud logs tail zeno-auth-dev --region=europe-west3
```

**Common causes:**
1. Database connection failed
2. Migrations didn't run
3. Port mismatch (should be 8080)

---

## 📝 Post-Deployment Tasks

### Immediate (After Deploy)
- [ ] Verify health endpoints return 200
- [ ] Check Cloud Logging for errors
- [ ] Test user registration
- [ ] Test user login

### Test API:
```bash
# Register user
curl -X POST $SERVICE_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#",
    "full_name": "Test User"
  }'

# Login
curl -X POST $SERVICE_URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#"
  }'
```

---

## 🔄 Update Deployment

To deploy a new version:

```bash
# Make your code changes
git add .
git commit -m "Your changes"

# Redeploy
./gcp-deploy.sh
```

Cloud Run will:
1. Build new image with timestamp tag
2. Deploy new revision
3. Gradually shift traffic to new revision
4. Keep old revision for rollback

---

## ↩️ Rollback

If something goes wrong:

```bash
# List revisions
gcloud run revisions list --service=zeno-auth-dev --region=europe-west3

# Rollback to previous revision
gcloud run services update-traffic zeno-auth-dev \
  --region=europe-west3 \
  --to-revisions=PREVIOUS_REVISION_NAME=100

# Verify
curl $SERVICE_URL/health
```

---

## 📊 Monitoring

### View Logs
```bash
# Real-time
gcloud logs tail zeno-auth-dev --region=europe-west3

# Last 100 entries
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=100

# Filter errors only
gcloud logs read zeno-auth-dev --region=europe-west3 \
  --filter="severity>=ERROR" --limit=50
```

### Service Metrics
```bash
# Open in Cloud Console
echo "https://console.cloud.google.com/run/detail/europe-west3/zeno-auth-dev/metrics"
```

### Custom Metrics
```bash
# Prometheus metrics endpoint (protected)
curl $SERVICE_URL/metrics
```

---

## 🧹 Cleanup (Delete Everything)

**⚠️ WARNING: This will delete all resources!**

```bash
# Delete Cloud Run service
gcloud run services delete zeno-auth-dev --region=europe-west3 --quiet

# Delete secrets
gcloud secrets delete zeno-auth-database-url --quiet
gcloud secrets delete zeno-auth-jwt-private-key --quiet

# Delete service account
gcloud iam service-accounts delete \
  zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com --quiet

# Delete Artifact Registry images (optional)
gcloud artifacts repositories delete zeno-auth \
  --location=europe-west3 --quiet
```

---

## 📚 Additional Documentation

- **[QUICK_DEPLOY.md](./QUICK_DEPLOY.md)** - Quick start guide
- **[../GCP_PRODUCTION_CHECKLIST.md](../GCP_PRODUCTION_CHECKLIST.md)** - Detailed checklist
- **[../DEPLOY_STATUS.md](../DEPLOY_STATUS.md)** - Current deployment status
- **[../FIXES_APPLIED.md](../FIXES_APPLIED.md)** - What was fixed
- **[GCP_DEPLOYMENT.md](./GCP_DEPLOYMENT.md)** - Full deployment guide

---

## 🆘 Need Help?

1. **Check logs:** `gcloud logs read zeno-auth-dev --limit=100`
2. **Verify secrets:** `gcloud secrets list | grep zeno-auth`
3. **Check IAM:** `./setup-iam.sh` (will show current roles)
4. **Run diagnostics:** `./pre-deploy-check.sh`
5. **Review checklist:** [GCP_PRODUCTION_CHECKLIST.md](../GCP_PRODUCTION_CHECKLIST.md)

---

## ✅ Success Criteria

Deployment is successful when:

1. ✅ Service shows "Healthy" in Cloud Console
2. ✅ `curl $SERVICE_URL/health` returns `{"status":"alive"}`
3. ✅ `curl $SERVICE_URL/health/ready` returns `{"db":"up"}`
4. ✅ No errors in Cloud Logging
5. ✅ Can register and login test user

---

**Ready to deploy?**

```bash
./gcp-deploy.sh
```

🚀 Good luck!
