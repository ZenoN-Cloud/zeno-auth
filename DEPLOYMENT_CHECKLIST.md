# Deployment Checklist for zeno-auth

## ✅ Pre-Deployment

- [ ] Code review completed
- [ ] All tests passing
- [ ] JWT keys generated (`make generate-keys`)
- [ ] Environment variables documented
- [ ] Database migrations tested
- [ ] Dockerfile builds successfully
- [ ] .gcloudignore configured

## ✅ GCP Setup

- [ ] GCP project created
- [ ] Billing enabled
- [ ] Required APIs enabled:
  - [ ] Cloud Run API
  - [ ] Cloud Build API
  - [ ] Secret Manager API
  - [ ] Cloud SQL Admin API
  - [ ] Container Registry API
  
- [ ] Cloud SQL instance created
- [ ] Database `zeno_auth` created
- [ ] Database user created with strong password
- [ ] Cloud SQL Proxy configured (if needed)

## ✅ Secrets Configuration

- [ ] `zeno-auth-database-url` secret created
- [ ] `zeno-auth-jwt-private-key` secret created
- [ ] Service account has access to secrets
- [ ] Secrets tested and validated

## ✅ Deployment

- [ ] `GCP_PROJECT_ID` environment variable set
- [ ] `GCP_REGION` environment variable set
- [ ] Run `./deploy-cloudrun.sh`
- [ ] Verify deployment: `curl <SERVICE_URL>/health`
- [ ] Check logs for errors
- [ ] Test authentication endpoints

## ✅ Post-Deployment

- [ ] Health check endpoint responding
- [ ] JWKS endpoint accessible
- [ ] Database migrations applied
- [ ] Monitor logs for errors
- [ ] Set up monitoring/alerting
- [ ] Document service URL
- [ ] Update frontend configuration with new auth URL

## 🔧 Rollback Plan

If deployment fails:
```bash
# Get previous revision
gcloud run revisions list --service zeno-auth --region europe-west1

# Rollback to previous revision
gcloud run services update-traffic zeno-auth \
  --to-revisions=<PREVIOUS_REVISION>=100 \
  --region europe-west1
```

## 📊 Monitoring

```bash
# View logs
gcloud run logs read zeno-auth --region europe-west1 --limit 100

# Check service status
gcloud run services describe zeno-auth --region europe-west1

# View metrics
gcloud monitoring dashboards list
```
