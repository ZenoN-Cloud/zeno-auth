# Cleanup Cron Job Setup

## Overview

The cleanup job removes expired data according to GDPR data retention policies:
- Expired refresh tokens (older than 24 hours after expiration)
- Old audit logs (default: 730 days / 2 years)
- Expired email verification tokens (7 days after expiration)
- Expired password reset tokens (7 days after expiration)

## Local Development

Run manually:
```bash
./scripts/run-cleanup.sh
```

With custom retention:
```bash
./scripts/run-cleanup.sh 365  # 1 year retention
```

## Production Setup

### Option 1: Cron (Linux/Unix)

Add to crontab:
```bash
# Run cleanup daily at 2 AM
0 2 * * * cd /path/to/zeno-auth && ./scripts/run-cleanup.sh >> /var/log/zeno-auth-cleanup.log 2>&1
```

Edit crontab:
```bash
crontab -e
```

### Option 2: GCP Cloud Scheduler

1. Build Docker image with cleanup command:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o cleanup ./cmd/cleanup/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/cleanup .
CMD ["./cleanup", "-retention-days=730"]
```

2. Create Cloud Run job:
```bash
gcloud run jobs create zeno-auth-cleanup \
  --image gcr.io/PROJECT_ID/zeno-auth-cleanup:latest \
  --region us-central1 \
  --set-env-vars DATABASE_URL=postgresql://... \
  --max-retries 3
```

3. Create Cloud Scheduler:
```bash
gcloud scheduler jobs create http zeno-auth-cleanup-schedule \
  --location us-central1 \
  --schedule "0 2 * * *" \
  --uri "https://us-central1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/PROJECT_ID/jobs/zeno-auth-cleanup:run" \
  --http-method POST \
  --oauth-service-account-email SERVICE_ACCOUNT@PROJECT_ID.iam.gserviceaccount.com
```

### Option 3: Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: zeno-auth-cleanup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: gcr.io/PROJECT_ID/zeno-auth-cleanup:latest
            env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: zeno-auth-secrets
                  key: database-url
            args: ["-retention-days=730"]
          restartPolicy: OnFailure
```

## Monitoring

Check logs:
```bash
# Local
tail -f /var/log/zeno-auth-cleanup.log

# GCP Cloud Run
gcloud logging read "resource.type=cloud_run_job AND resource.labels.job_name=zeno-auth-cleanup" --limit 50

# Kubernetes
kubectl logs -l job-name=zeno-auth-cleanup
```

## Retention Policies

| Data Type | Retention | Reason |
|-----------|-----------|--------|
| Audit logs | 730 days (2 years) | GDPR Art. 30 - Legal requirement |
| Refresh tokens | 90 days after revocation | Security best practice |
| Email verification tokens | 7 days after expiration | Cleanup old tokens |
| Password reset tokens | 7 days after expiration | Cleanup old tokens |

## Customization

Modify retention in `cmd/cleanup/main.go`:
```go
retentionDays := flag.Int("retention-days", 730, "Audit log retention in days")
```

Or pass as argument:
```bash
./cleanup -retention-days=365
```
