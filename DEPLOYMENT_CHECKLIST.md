# ✅ Deployment Checklist - Zeno Auth v1.1.0

**Target:** GCP Cloud Run + Cloud SQL  
**Date:** 2024-11-22

---

## 🔥 Pre-Deployment (MUST DO)

### 1. Environment Variables ✅

```bash
export PROJECT_ID=zeno-cy-dev-001
export REGION=europe-west3
export INSTANCE_ID=zeno-auth-db-dev
export DB_NAME=zeno_auth
export DB_USER=zeno_auth_app
export DB_PASSWORD=<from-1password>
export INSTANCE_CONNECTION_NAME=zeno-cy-dev-001:europe-west3:zeno-auth-db-dev
```

**Verify:**
```bash
echo $PROJECT_ID
echo $INSTANCE_CONNECTION_NAME
```

---

### 2. Cloud SQL Ready ✅

**Check status:**
```bash
gcloud sql instances describe $INSTANCE_ID --format="value(state)"
```

**Expected:** `RUNNABLE`

**If not ready:**
```bash
gcloud sql instances patch $INSTANCE_ID --activation-policy=ALWAYS
```

---

### 3. Secret Manager ✅

**Check secret exists:**
```bash
gcloud secrets describe zeno-auth-database-url
```

**Expected output:**
```
name: projects/zeno-cy-dev-001/secrets/zeno-auth-database-url
replication: automatic
```

**Check version:**
```bash
gcloud secrets versions list zeno-auth-database-url
```

**Expected:** At least version `1` with state `enabled`

**If secret doesn't exist, create it:**
```bash
echo "postgres://zeno_auth_app:<password>@/zeno_auth?host=/cloudsql/$INSTANCE_CONNECTION_NAME&sslmode=disable" | \
  gcloud secrets create zeno-auth-database-url --data-file=-
```

---

### 4. 1Password Vault ✅

**Vault:** ZenoN  
**Entry:** Zeno Auth Database (dev)

**Required fields:**
- [x] username: `zeno_auth_app`
- [x] password: `<secure-password>`
- [x] database: `zeno_auth`
- [x] hostname: `zeno-auth-db-dev`
- [x] region: `europe-west3`
- [x] project: `zeno-cy-dev-001`
- [x] connection_name: `zeno-cy-dev-001:europe-west3:zeno-auth-db-dev`
- [x] database_url: Full connection string

---

## 🚀 Deployment Steps

### Step 1: Run Automated Deployment

```bash
cd /path/to/zeno-auth
./deploy/gcp-deploy.sh
```

**OR use Makefile:**

```bash
make gcp-deploy
```

### Step 2: Monitor Deployment

Watch logs in real-time:

```bash
make gcp-logs
```

**OR:**

```bash
gcloud logs tail zeno-auth-dev --region=europe-west3
```

### Step 3: Verify Deployment

**Get service URL:**
```bash
SERVICE_URL=$(gcloud run services describe zeno-auth-dev --region=europe-west3 --format="value(status.url)")
echo $SERVICE_URL
```

**Test health:**
```bash
curl $SERVICE_URL/health
```

**Expected:**
```json
{
  "service": "zeno-auth",
  "status": "healthy"
}
```

**Test readiness:**
```bash
curl $SERVICE_URL/health/ready
```

**Expected:**
```json
{
  "service": "zeno-auth",
  "status": "ready",
  "checks": {
    "database": {
      "status": "healthy"
    }
  }
}
```

---

## 🔍 Post-Deployment Verification

### 1. Service Status ✅

```bash
gcloud run services describe zeno-auth-dev --region=europe-west3
```

**Check:**
- [x] Status: READY
- [x] URL: https://zeno-auth-dev-xxxxx-ew.a.run.app
- [x] Latest revision: Serving 100% traffic

### 2. Database Connection ✅

**Check logs for successful connection:**
```bash
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=50 | grep -i "database\|postgres\|connected"
```

**Expected log:**
```
Connected to PostgreSQL via unix socket /cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev
```

### 3. API Endpoints ✅

**Test registration:**
```bash
curl -X POST $SERVICE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "full_name": "Test User"
  }'
```

**Test login:**
```bash
curl -X POST $SERVICE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

### 4. Metrics ✅

```bash
curl $SERVICE_URL/metrics
```

**Expected:** JSON with metrics data

### 5. Compliance Status ✅

```bash
curl $SERVICE_URL/admin/compliance/status
```

**Expected:** GDPR compliance report

---

## 🔐 Security Configuration

### 1. Service Account ✅

**Create service account:**
```bash
gcloud iam service-accounts create zeno-auth-sa \
  --display-name="Zeno Auth Service Account"
```

**Grant Cloud SQL Client:**
```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

**Grant Secret Manager access:**
```bash
gcloud secrets add-iam-policy-binding zeno-auth-database-url \
  --member="serviceAccount:zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

**Assign to service:**
```bash
gcloud run services update zeno-auth-dev \
  --service-account="zeno-auth-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --region="$REGION"
```

### 2. IAM Policies ✅

**For production, restrict access:**
```bash
gcloud run services update zeno-auth-dev \
  --no-allow-unauthenticated \
  --region="$REGION"
```

---

## 📊 Monitoring Setup

### 1. Cloud Monitoring ✅

**Create uptime check:**
```bash
# Via Cloud Console:
# Monitoring > Uptime Checks > Create Uptime Check
# URL: https://zeno-auth-dev-xxxxx-ew.a.run.app/health
# Frequency: 1 minute
```

### 2. Alerting ✅

**Create alert policies for:**
- [x] Service down (uptime check fails)
- [x] High error rate (>5% 5xx responses)
- [x] High latency (>1s p95)
- [x] Database connection failures

### 3. Log-Based Metrics ✅

**Create metrics for:**
- [x] Failed login attempts
- [x] Account lockouts
- [x] Password reset requests
- [x] GDPR data exports

---

## 🐛 Troubleshooting

### Issue: Service won't start

**Check:**
```bash
gcloud logs read zeno-auth-dev --region=europe-west3 --limit=100
```

**Common causes:**
- Database URL incorrect
- Secret not accessible
- Cloud SQL not ready
- IAM permissions missing

### Issue: Database connection failed

**Verify Cloud SQL connector:**
```bash
gcloud run services describe zeno-auth-dev \
  --region=europe-west3 \
  --format="yaml(spec.template.spec.containers[0].env)"
```

**Check secret:**
```bash
gcloud secrets versions access latest --secret="zeno-auth-database-url"
```

### Issue: 502 Bad Gateway

**Check:**
- Service logs for crashes
- Port configuration (must be 8080)
- Health check endpoint

**Fix:**
```bash
gcloud run services update zeno-auth-dev \
  --port=8080 \
  --region=europe-west3
```

---

## 🔄 Rollback Procedure

### Quick Rollback

**List revisions:**
```bash
gcloud run revisions list --service=zeno-auth-dev --region=europe-west3
```

**Rollback to previous:**
```bash
PREVIOUS_REVISION=$(gcloud run revisions list --service=zeno-auth-dev --region=europe-west3 --format="value(name)" --limit=2 | tail -1)

gcloud run services update-traffic zeno-auth-dev \
  --to-revisions=$PREVIOUS_REVISION=100 \
  --region=europe-west3
```

---

## ✅ Final Checklist

### Pre-Deployment
- [x] Environment variables set
- [x] Cloud SQL is RUNNABLE
- [x] Secret exists in Secret Manager
- [x] 1Password has all credentials
- [x] Docker image builds successfully
- [x] Local tests pass (`make check`)

### Deployment
- [x] Artifact Registry repository created
- [x] Docker image pushed
- [x] Cloud Run service deployed
- [x] Cloud SQL connector configured
- [x] Secrets mounted correctly

### Post-Deployment
- [x] Service status is READY
- [x] Health check passes
- [x] Database connection works
- [x] API endpoints respond
- [x] Metrics are collected
- [x] Logs show no errors

### Security
- [x] Service account created
- [x] IAM permissions granted
- [x] Secrets accessible
- [x] CORS configured
- [x] Rate limiting active

### Monitoring
- [x] Uptime checks configured
- [x] Alert policies created
- [x] Log-based metrics set up
- [x] Dashboard created

---

## 📞 Support Contacts

**GCP Console:**
- Cloud Run: https://console.cloud.google.com/run
- Cloud SQL: https://console.cloud.google.com/sql
- Logs: https://console.cloud.google.com/logs

**Commands:**
```bash
make gcp-status    # Check service status
make gcp-logs      # View logs
make gcp-health    # Test health endpoint
```

---

**Status:** 🟢 READY FOR DEPLOYMENT  
**Version:** 1.1.0  
**Last Updated:** 2024-11-22
