# 🚀 Quick Deploy to GCP Cloud Run

**Time to deploy:** ~15 minutes  
**Prerequisites:** gcloud CLI, Docker, Cloud SQL instance

---

## 📋 Pre-Flight Checklist

Before running deployment, ensure:

- [ ] Cloud SQL instance `zeno-auth-db-dev` is **RUNNABLE**
- [ ] Database `zeno_auth` created
- [ ] User `zeno_auth_app` created with password
- [ ] gcloud CLI authenticated: `gcloud auth login`
- [ ] Docker running locally

---

## 🎯 Quick Start (3 Commands)

### 1️⃣ Setup Secrets & IAM (First Time Only)

```bash
cd deploy
./gcp-setup-secrets.sh
```

**This script will:**
- ✅ Create/verify service account `zeno-auth-sa`
- ✅ Grant IAM roles (Cloud SQL, Secret Manager, Logging)
- ✅ Create `DATABASE_URL` secret (you'll be prompted)
- ✅ Generate JWT keys and store in Secret Manager

**You'll need:**
- DATABASE_URL connection string (see format below)

---

### 2️⃣ Deploy to Cloud Run

```bash
./gcp-deploy.sh
```

**This script will:**
- ✅ Build Docker image
- ✅ Push to Artifact Registry
- ✅ Deploy to Cloud Run
- ✅ Run health check

**Duration:** ~5-7 minutes

---

### 3️⃣ Verify Deployment

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe zeno-auth-dev \
  --region=europe-west3 \
  --format="value(status.url)")

# Test health endpoint
curl $SERVICE_URL/health

# Expected response:
# {"status":"alive","timestamp":"..."}

# Test readiness (includes DB check)
curl $SERVICE_URL/health/ready

# Expected response:
# {"status":"ready","db":"up","timestamp":"..."}
```

---

## 🔐 DATABASE_URL Format

### Option 1: Unix Socket (Recommended)

```
postgres://zeno_auth_app:YOUR_PASSWORD@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev&sslmode=disable
```

**Replace:**
- `YOUR_PASSWORD` - your database password
- `zeno-cy-dev-001` - your project ID
- `europe-west3` - your region
- `zeno-auth-db-dev` - your Cloud SQL instance name

### Option 2: Private IP (Advanced)

```
postgres://zeno_auth_app:YOUR_PASSWORD@10.0.0.5:5432/zeno_auth?sslmode=require
```

**Requires:**
- VPC Connector configured
- Private IP enabled on Cloud SQL

---

## 🔧 Configuration Variables

Edit `deploy/.env.gcp.example` or set environment variables:

```bash
export PROJECT_ID="zeno-cy-dev-001"
export REGION="europe-west3"
export INSTANCE_ID="zeno-auth-db-dev"
export SERVICE_NAME="zeno-auth-dev"
```

---

## 🐛 Troubleshooting

### Issue: "Secret not found"

```bash
# Check if secret exists
gcloud secrets list | grep zeno-auth

# Create DATABASE_URL secret manually
echo -n "postgres://..." | gcloud secrets create zeno-auth-database-url --data-file=-

# Verify
gcloud secrets versions access latest --secret=zeno-auth-database-url
```

### Issue: "Permission denied" on Cloud SQL

```bash
# Verify service account has roles
gcloud projects get-iam-policy zeno-cy-dev-001 \
  --flatten="bindings[].members" \
  --filter="bindings.members:zeno-auth-sa"

# Re-run setup script
./gcp-setup-secrets.sh
```

### Issue: "Migration failed"

```bash
# Check Cloud Run logs
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=50

# Common causes:
# - DATABASE_URL incorrect
# - Cloud SQL instance not connected
# - User permissions missing
```

### Issue: Health check fails

```bash
# Check service logs
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=50 --format=json

# Check Cloud SQL connection
gcloud sql instances describe zeno-auth-db-dev

# Test database connection locally
psql "$DATABASE_URL" -c "SELECT 1"
```

---

## 📊 Post-Deployment

### View Logs

```bash
# Real-time logs
gcloud logs tail zeno-auth-dev --region=europe-west3

# Last 50 entries
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=50
```

### Monitor Service

```bash
# Service details
gcloud run services describe zeno-auth-dev --region=europe-west3

# Open in console
echo "https://console.cloud.google.com/run/detail/europe-west3/zeno-auth-dev"
```

### Update Service

```bash
# Rebuild and redeploy
./gcp-deploy.sh

# Update environment variable
gcloud run services update zeno-auth-dev \
  --region=europe-west3 \
  --set-env-vars=NEW_VAR=value
```

---

## 🔄 Rollback

```bash
# List revisions
gcloud run revisions list --service=zeno-auth-dev --region=europe-west3

# Rollback to previous revision
gcloud run services update-traffic zeno-auth-dev \
  --region=europe-west3 \
  --to-revisions=zeno-auth-dev-00001-abc=100
```

---

## 🧹 Cleanup (Delete Everything)

```bash
# Delete Cloud Run service
gcloud run services delete zeno-auth-dev --region=europe-west3

# Delete secrets
gcloud secrets delete zeno-auth-database-url
gcloud secrets delete zeno-auth-jwt-private-key

# Delete service account
gcloud iam service-accounts delete zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com

# Delete Cloud SQL instance (CAREFUL!)
gcloud sql instances delete zeno-auth-db-dev
```

---

## 📚 Additional Resources

- **Full Checklist:** [GCP_PRODUCTION_CHECKLIST.md](../GCP_PRODUCTION_CHECKLIST.md)
- **GCP Deployment Guide:** [GCP_DEPLOYMENT.md](./GCP_DEPLOYMENT.md)
- **Environment Variables:** [../docs/ENV_VARIABLES.md](../docs/ENV_VARIABLES.md)
- **Architecture:** [../docs/architecture.md](../docs/architecture.md)

---

## 🎯 Success Criteria

Your deployment is successful when:

- ✅ `curl $SERVICE_URL/health` returns `{"status":"alive"}`
- ✅ `curl $SERVICE_URL/health/ready` returns `{"db":"up"}`
- ✅ No errors in Cloud Logging
- ✅ Service shows "Healthy" in Cloud Console
- ✅ Can register a test user via API

---

## 🆘 Need Help?

1. Check [GCP_PRODUCTION_CHECKLIST.md](../GCP_PRODUCTION_CHECKLIST.md) for detailed troubleshooting
2. Review Cloud Run logs: `gcloud logs read zeno-auth-dev --limit=100`
3. Verify all secrets exist: `gcloud secrets list | grep zeno-auth`
4. Check IAM roles: `./gcp-setup-secrets.sh` (Step 4)

---

**Last Updated:** 2024  
**Version:** 1.1.0
