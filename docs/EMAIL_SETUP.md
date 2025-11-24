# Email Setup Guide

Email configuration for Zeno Auth using SendGrid.

## SendGrid Setup

### 1. Create SendGrid Account

1. Go to [SendGrid](https://sendgrid.com/)
2. Sign up for free account (100 emails/day free)
3. Verify your email address

### 2. Create API Key

1. Go to **Settings** → **API Keys**
2. Click **Create API Key**
3. Name: `zeno-auth-production`
4. Permissions: **Full Access** (or **Mail Send** only)
5. Copy the API key (you won't see it again!)

### 3. Verify Sender Identity

**Option A: Single Sender Verification (Quick)**
1. Go to **Settings** → **Sender Authentication**
2. Click **Verify a Single Sender**
3. Enter your email (e.g., `noreply@zenon-cloud.com`)
4. Fill in the form and submit
5. Check your email and click verification link

**Option B: Domain Authentication (Recommended for Production)**
1. Go to **Settings** → **Sender Authentication**
2. Click **Authenticate Your Domain**
3. Follow DNS setup instructions
4. Add CNAME records to your domain DNS

### 4. Store API Key in GCP Secret Manager

```bash
# Create secret
echo -n "SG.your-api-key-here" | gcloud secrets create zeno-auth-sendgrid-api-key \
  --data-file=- \
  --replication-policy=automatic

# Grant access to service account
gcloud secrets add-iam-policy-binding zeno-auth-sendgrid-api-key \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### 5. Update Cloud Run Service

```bash
gcloud run services update zeno-auth-dev \
  --region=europe-west3 \
  --set-secrets=SENDGRID_API_KEY=zeno-auth-sendgrid-api-key:latest \
  --update-env-vars=EMAIL_FROM=noreply@zenon-cloud.com,EMAIL_FROM_NAME="ZenoN Cloud",APP_BASE_URL=https://zeno-cy-frontend-dev-001.storage.googleapis.com
```

## Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `SENDGRID_API_KEY` | SendGrid API key | Yes | `SG.xxx` |
| `EMAIL_FROM` | Sender email address | Yes | `noreply@em2292.zeno-cy.com` |
| `EMAIL_FROM_NAME` | Sender name | No | `ZenoN Cloud` |
| `APP_BASE_URL` | Frontend URL for links | Yes | `https://zeno-cy-frontend-dev-001.storage.googleapis.com` |

## Email Templates

### Verification Email
- **Subject:** Verify your email address
- **Link:** `{APP_BASE_URL}/verify-email?token={token}`
- **Expires:** 24 hours

### Password Reset Email
- **Subject:** Reset your password
- **Link:** `{APP_BASE_URL}/reset-password?token={token}`
- **Expires:** 1 hour

### Password Changed Email
- **Subject:** Your password has been changed
- **Action:** Security notification

### Account Lockout Email
- **Subject:** Account temporarily locked
- **Info:** Locked until timestamp

## Testing

### Local Testing

```bash
# Set environment variables
export SENDGRID_API_KEY="SG.your-test-key"
export EMAIL_FROM="test@example.com"
export APP_BASE_URL="http://localhost:3000"

# Start service
make local-up
```

### Test Email Sending

```bash
# Register user (triggers verification email)
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123",
    "full_name": "Test User"
  }'

# Check logs for email sending
docker-compose logs -f auth
```

## Monitoring

### SendGrid Dashboard

1. Go to **Activity** → **Email Activity**
2. View sent emails, opens, clicks, bounces
3. Monitor delivery rates

### Application Logs

```bash
# View email logs
gcloud logging read "resource.type=cloud_run_revision AND textPayload=~\"email\"" \
  --limit=50 \
  --format=json
```

## Troubleshooting

### Email Not Sending

1. **Check API Key**
   ```bash
   gcloud secrets versions access latest --secret=zeno-auth-sendgrid-api-key
   ```

2. **Check Sender Verification**
   - Ensure sender email is verified in SendGrid
   - Check spam folder

3. **Check Logs**
   ```bash
   gcloud logging read "resource.type=cloud_run_revision AND severity>=ERROR" --limit=20
   ```

### Common Errors

**Error: `403 Forbidden`**
- API key is invalid or expired
- Regenerate API key in SendGrid

**Error: `sender identity not verified`**
- Sender email not verified in SendGrid
- Complete sender verification

**Error: `rate limit exceeded`**
- Free tier: 100 emails/day
- Upgrade SendGrid plan

## Production Checklist

- [ ] SendGrid account created
- [ ] API key generated and stored in Secret Manager
- [ ] Sender email verified (or domain authenticated)
- [ ] Environment variables configured in Cloud Run
- [ ] Test email sending works
- [ ] Monitor email delivery rates
- [ ] Set up email templates (optional)
- [ ] Configure unsubscribe handling (if needed)

## Cost

**SendGrid Free Tier:**
- 100 emails/day
- Basic email analytics
- Single sender verification

**Paid Plans:**
- Essentials: $19.95/month (50,000 emails)
- Pro: $89.95/month (100,000 emails)
- Custom pricing for higher volumes

## Alternative Providers

If SendGrid doesn't work for you:

1. **AWS SES** - $0.10 per 1,000 emails
2. **Mailgun** - 5,000 emails/month free
3. **Postmark** - 100 emails/month free
4. **Resend** - 3,000 emails/month free

---

**Last Updated:** 2024  
**Version:** 1.1.0
