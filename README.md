# 🔐 Zeno Auth

Core authentication and identity service for the **ZenoN-Cloud** platform.

[![Pipeline](https://gitlab.com/zeno-cy/zeno-auth/badges/main/pipeline.svg)](https://gitlab.com/zeno-cy/zeno-auth/-/pipelines)
[![Coverage](https://gitlab.com/zeno-cy/zeno-auth/badges/main/coverage.svg)](https://gitlab.com/zeno-cy/zeno-auth/-/commits/main)
[![Production Ready](https://img.shields.io/badge/status-production%20ready-brightgreen)](docs/architecture.md)
[![GDPR Compliant](https://img.shields.io/badge/GDPR-compliant-blue)](docs/GDPR_COMPLIANCE.md)

## 🎯 Features

- ✅ User registration & JWT authentication
- ✅ Password reset & email verification
- ✅ Session & organization management
- ✅ Argon2id hashing & rate limiting
- ✅ GDPR compliance (Art. 15, 17, 20)
- ✅ Prometheus metrics & health checks

## 🚀 Quick Start

### Local Development

```bash
# Start services
docker-compose up -d

# Run tests
make test

# Check health
curl http://localhost:8080/health
```

**Access:**

- API: http://localhost:8080
- pgAdmin: http://localhost:5050

### GCP Deployment

```bash
# Setup infrastructure (one-time)
make gcp-setup

# Deploy application
make gcp-deploy

# Check status
make gcp-status-check
```

## 📚 Documentation

- **[Architecture](docs/architecture.md)** - System design & components
- **[GDPR Compliance](docs/GDPR_COMPLIANCE.md)** - Privacy & data protection
- **[Environment Variables](docs/ENV_VARIABLES.md)** - Configuration reference
- **[Password Policy](docs/PASSWORD_POLICY.md)** - Security requirements
- **[Email Setup](docs/EMAIL_SETUP.md)** - SendGrid configuration
- **[Deployment Guide](deploy/README.md)** - Production deployment
- **[GitLab CI/CD Setup](.gitlab/GITLAB_SETUP.md)** - CI/CD configuration
- **[API Specification](api/openapi.yaml)** - OpenAPI 3.0 spec
- **[Changelog](CHANGELOG.md)** - Version history

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
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

**Stack:**

- Go 1.25 + Gin
- PostgreSQL 17
- JWT (RS256)
- Argon2id
- Zerolog

## 🔧 Development

```bash
# Format & test
make check

# Build
make build

# Coverage
make cover

# Integration tests
make local-up
make integration
```

## 🚢 Deployment

### Environment Variables

```env
DATABASE_URL=postgres://user:pass@host:5432/dbname
JWT_PRIVATE_KEY=<base64-encoded-key>
CORS_ALLOWED_ORIGINS=https://storage.googleapis.com
ENV=production
PORT=8080
```

### GCP Cloud Run

```bash
# Check infrastructure
make gcp-status-check

# Deploy
make gcp-deploy

# View logs
make gcp-logs

# Test
make gcp-test
```

## 📊 Monitoring

```bash
# Metrics
curl http://localhost:8080/metrics

# Health checks
curl http://localhost:8080/health
curl http://localhost:8080/health/ready
curl http://localhost:8080/health/live
```

## 🔐 Security

- ✅ Argon2id password hashing
- ✅ Rate limiting (5 attempts / 15 min)
- ✅ Account lockout after 5 failed logins
- ✅ Session fingerprinting
- ✅ Security headers (HSTS, CSP, etc.)
- ✅ Input validation & sanitization
- ✅ Audit logging

## 📝 API Endpoints

### Authentication

- `POST /v1/auth/register` - Register user
- `POST /v1/auth/login` - Login
- `POST /v1/auth/refresh` - Refresh token
- `POST /v1/auth/logout` - Logout
- `POST /v1/auth/forgot-password` - Request reset
- `POST /v1/auth/reset-password` - Reset password

### User

- `GET /v1/me` - Get profile
- `POST /v1/me/change-password` - Change password
- `GET /v1/me/sessions` - List sessions
- `DELETE /v1/me/sessions/:id` - Revoke session

### GDPR

- `GET /v1/me/data-export` - Export data (Art. 15)
- `DELETE /v1/me/account` - Delete account (Art. 17)
- `GET /v1/me/consents` - List consents
- `POST /v1/me/consents` - Grant consent
- `DELETE /v1/me/consents/:type` - Revoke consent

### Health

- `GET /health` - Basic health
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe
- `GET /metrics` - Prometheus metrics
- `GET /.well-known/jwks.json` - JWKS

See [OpenAPI spec](api/openapi.yaml) for full documentation.

## 🧪 Testing

```bash
# Unit tests
make test

# With coverage
make cover

# Integration tests
make integration

# E2E tests
E2E_BASE_URL=http://localhost:8080 make e2e
```

## 📦 Project Structure

```
zeno-auth/
├── cmd/
│   ├── auth/          # Main service
│   └── cleanup/       # Cleanup job
├── internal/
│   ├── handler/       # HTTP handlers
│   ├── service/       # Business logic
│   ├── repository/    # Data access
│   ├── model/         # Domain models
│   └── token/         # JWT & crypto
├── migrations/        # SQL migrations
├── deploy/            # Deployment scripts
├── docs/              # Documentation
└── test/              # Integration tests
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open Merge Request

See [Merge Request Template](.gitlab/merge_request_templates/Default.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**Repository:** [gitlab.com/zeno-cy/zeno-auth](https://gitlab.com/zeno-cy/zeno-auth)  
**Status:** 🟢 Production Ready  
**Version:** 1.1.0  
**Deployed:** https://zeno-auth-dev-umu7aajgeq-ey.a.run.app
