# 🔐 Zeno Auth

Core authentication and identity service for the **ZenoN-Cloud** platform.

[![Production Ready](https://img.shields.io/badge/status-production%20ready-brightgreen)](docs/IMPLEMENTATION_STATUS.md)
[![GDPR Compliant](https://img.shields.io/badge/GDPR-compliant-blue)](docs/GDPR_COMPLIANCE.md)
[![Security Score](https://img.shields.io/badge/security-86%25-green)](docs/IMPLEMENTATION_STATUS.md)

## 🎯 Features

### Core Authentication
- ✅ User registration & login
- ✅ JWT access & refresh tokens
- ✅ Password reset flow
- ✅ Email verification
- ✅ Session management
- ✅ Organization management

### Security
- ✅ Argon2id password hashing
- ✅ Rate limiting (brute-force protection)
- ✅ Session fingerprinting
- ✅ Account lockout (5 failed attempts)
- ✅ Input validation & sanitization
- ✅ Security headers (HSTS, CSP, etc.)
- ✅ CORS whitelist

### GDPR Compliance
- ✅ Right to Access (Art. 15)
- ✅ Right to Erasure (Art. 17)
- ✅ Right to Data Portability (Art. 20)
- ✅ Consent Management (Art. 7)
- ✅ Data Retention Policy (Art. 5.1.e)
- ✅ Audit Logging (Art. 30)
- ✅ Privacy by Design (Art. 25)

### Production Features
- ✅ Prometheus metrics
- ✅ Enhanced health checks (liveness/readiness)
- ✅ Structured logging (Zerolog)
- ✅ OpenAPI documentation
- ✅ Admin panel with compliance reports
- ✅ Automated cleanup jobs
## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+ (for local development)

### Start All Services

```bash
docker-compose up -d
```

**Access:**
- 🎨 **Frontend:** http://localhost:5173
- 🔌 **API:** http://localhost:8080
- 📊 **Admin Panel:** http://localhost:5173 (click "Admin" button)
- 🗄️ **pgAdmin:** http://localhost:5050

### Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Readiness probe
curl http://localhost:8080/health/ready

# Metrics
curl http://localhost:8080/metrics

# Compliance status
curl http://localhost:8080/admin/compliance/status
```

## 📚 Documentation

### 🚀 Quick Start
- **[QUICKSTART.md](./QUICKSTART.md)** - Quick start guide
- **[QUICK_REFERENCE.md](./QUICK_REFERENCE.md)** - ⭐ **Quick reference for developers**
- **[FULL_STACK_LOCAL.md](./FULL_STACK_LOCAL.md)** - Full stack local setup
- **[LOCAL_DEV.md](./LOCAL_DEV.md)** - Local development guide

### 🏗️ Architecture & Implementation

- **[REFACTORING_COMPLETE.md](./REFACTORING_COMPLETE.md)** - ⭐ **Latest refactoring (v1.1.0)**
- **[ARCHITECTURE_IMPROVEMENTS.md](./ARCHITECTURE_IMPROVEMENTS.md)** - Architecture improvements checklist
- **[IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md)** - Detailed implementation summary
- **[NEXT_STEPS.md](./NEXT_STEPS.md)** - ⭐ **Next steps and priorities**
- **[docs/architecture.md](./docs/architecture.md)** - Service architecture
- **[docs/IMPLEMENTATION_STATUS.md](./docs/IMPLEMENTATION_STATUS.md)** - Implementation checklist

### 🔐 Security & Compliance
- **[docs/GDPR_COMPLIANCE.md](./docs/GDPR_COMPLIANCE.md)** - GDPR compliance documentation
- **[docs/SECURITY_FEATURES.md](./docs/SECURITY_FEATURES.md)** - Security features overview
- **[docs/PASSWORD_POLICY.md](./docs/PASSWORD_POLICY.md)** - ⭐ **Password policy and requirements**
- **[SECURITY_CHECKLIST.md](./SECURITY_CHECKLIST.md)** - Security & deployment checklist

### ⚙️ Configuration & Operations

- **[docs/ENV_VARIABLES.md](./docs/ENV_VARIABLES.md)** - ⭐ **Environment variables documentation**
- **[docs/CLEANUP_CRON.md](./docs/CLEANUP_CRON.md)** - Data retention & cleanup
- **[deploy/README.md](./deploy/README.md)** - Production deployment
- **[api/openapi.yaml](./api/openapi.yaml)** - OpenAPI specification

## 🏗️ Architecture

```
┌─────────────┐
│   Frontend  │ (React + TypeScript)
│  Port 5173  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Zeno Auth  │ (Go + Gin)
│  Port 8080  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ PostgreSQL  │
│  Port 5432  │
└─────────────┘
```

### Tech Stack

**Backend:**
- Go 1.25
- Gin (HTTP framework)
- PostgreSQL 17
- JWT (golang-jwt/jwt)
- Argon2id (password hashing)
- Zerolog (structured logging)

**Frontend:**
- React 18
- TypeScript
- Vite
- Yarn

## 📊 Status

**Implementation Progress:** 25/25 features (100%)  
**GDPR Compliance:** 10/10 (100%)  
**Security Score:** 13/14 (93%)  
**Production Ready:** ✅ Yes

See [IMPLEMENTATION_STATUS.md](./docs/IMPLEMENTATION_STATUS.md) for detailed breakdown.

## 🔧 Development

### Run Tests

```bash
go test ./... -v
```

### Format Code

```bash
go fmt ./...
```

### Lint Code

```bash
go vet ./...
```

### Build

```bash
go build -o auth ./cmd/auth
```

### Run Cleanup Job

```bash
./scripts/run-cleanup.sh
```

## 🚢 Deployment

See [deploy/README.md](./deploy/README.md) for production deployment instructions.

### Environment Variables

```env
DATABASE_URL=postgres://user:pass@host:5432/dbname
JWT_PRIVATE_KEY=<base64-encoded-private-key>
JWT_PUBLIC_KEY=<base64-encoded-public-key>
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
ENV=production
```

## 📈 Monitoring

### Metrics Endpoint

```bash
curl http://localhost:8080/metrics
```

**Available Metrics:**
- `auth_registrations_total`
- `auth_logins_total`
- `auth_login_failures_total`
- `auth_token_refreshes_total`
- `auth_active_sessions`
- `auth_request_duration_seconds` (histogram)

### Health Checks

```bash
# Basic health
curl http://localhost:8080/health

# Readiness (includes DB check)
curl http://localhost:8080/health/ready

# Liveness (system metrics)
curl http://localhost:8080/health/live
```

## 🔐 Security

### Implemented
- ✅ Argon2id password hashing
- ✅ Rate limiting (5 login attempts / 15 min)
- ✅ Session fingerprinting
- ✅ Account lockout after 5 failed attempts
- ✅ Input validation & sanitization
- ✅ Security headers (HSTS, CSP, X-Frame-Options, etc.)
- ✅ CORS whitelist
- ✅ Audit logging
- ✅ SQL injection prevention (parameterized queries)
- ✅ XSS prevention

### Production Hardening ✅
- ✅ Centralized error handling
- ✅ Non-root Docker user
- ✅ Fail-fast migrations
- ✅ Stdout-only logging in production
- ✅ Protected /metrics and /debug endpoints
- ✅ golangci-lint in CI
- ✅ Security test suite

### Recent Improvements (2024)

- ✅ **Fixed:** Account lockout now works correctly
- ✅ **Improved:** Session fingerprinting uses full IP address
- ✅ **Added:** Email notifications for critical events (4 types)
- ✅ **Enhanced:** GDPR compliance to 100%

### TODO
- ⏳ MFA/2FA (TOTP)
- ⏳ Email provider integration (SendGrid/AWS SES)
- ⏳ Encryption at rest

## 📝 License

MIT License - see [LICENSE](./LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please read the implementation status and security guidelines before contributing.

## 📞 Support

For issues and questions, please open a GitHub issue.

---

**Status:** 🟢 Production Ready  
**Last Updated:** 2024  
**Version:** 1.1.0
