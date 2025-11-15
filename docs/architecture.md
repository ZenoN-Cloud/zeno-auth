# 📘 Zeno Auth — Architecture (Revised, Production Grade)

Core authentication and identity service for the **ZenoN-Cloud** platform.
This document serves as the authoritative reference for:

* architecture & platform boundaries
* data model
* token model
* deployment model (GCP: Cloud Run, Cloud SQL, Secret Manager)
* service accounts & access control
* storage model
* CI/CD pipeline (GitHub Actions)
* environment strategy (dev → prod)
* security posture
* future expansion plans

This specification must be kept up to date whenever platform architecture evolves.

---

# 1. Purpose & Scope

`zeno-auth` provides:

* authentication
* identity
* organization management
* authorization metadata
* token issuing (JWT + refresh)
* introspection & identity boundaries

It is the **central identity provider** for the entire ZenoN-Cloud ecosystem.

It is *not* a gateway, document processor, billing engine, or storage system.

---

# 2. Environment Overview

The ZenoN-Cloud platform uses a strict separation between:

* **zenon-cloud-dev-001** → developer/staging infrastructure
* **zenon-cloud-prod-001** → production infrastructure

Both environments mirror the same structure:

```
Cloud SQL (Postgres 17)
│
├─ zeno_auth (database)
├─ zeno-docs buckets
└─ zeno-id-docs buckets
```

Service Accounts exist separately per environment:

```
zeno-auth-dev
zeno-docs-dev
zeno-id-docs-dev

zeno-auth-prod
zeno-docs-prod
zeno-id-docs-prod
```

Secrets exist separately per environment:

```
zeno-auth-db-dsn (dev)
zeno-auth-db-dsn (prod)
```

---

# 3. Deployment Model (GCP)

### 3.1 Runtime

`zeno-auth` runs on **Cloud Run** (fully serverless):

* auto-scaling
* minimal ops overhead
* integrates with Cloud SQL via Unix Sockets
* pulls secrets from Secret Manager
* uses service accounts to enforce access boundaries

### 3.2 Required GCP Services

* Artifact Registry (`zenon-cloud-dev-001` / `zenon-cloud-prod-001`)
* Cloud Run
* Secret Manager
* Cloud SQL Admin API
* IAM
* Cloud Logging

All services must be explicitly enabled.

---

# 4. Data Model

### 4.1 Users Table

(unchanged, but strictly enforced through migrations)

```
users:
  id UUID (PK)
  email TEXT unique, lowercased
  password_hash TEXT
  full_name TEXT
  is_active BOOLEAN
  created_at TIMESTAMP
  updated_at TIMESTAMP
```

### 4.2 Organizations Table

```
organizations:
  id UUID (PK)
  name TEXT
  owner_user_id UUID (FK → users.id)
  status ENUM(active, trial, suspended)
  created_at TIMESTAMP
  updated_at TIMESTAMP
```

### 4.3 org_memberships Table

```
org_memberships:
  id UUID
  user_id UUID
  org_id UUID
  role ENUM(OWNER, ADMIN, MEMBER, VIEWER)
  is_active BOOLEAN
  created_at TIMESTAMP
  UNIQUE(user_id, org_id)
```

### 4.4 refresh_tokens Table

```
refresh_tokens:
  id UUID
  user_id UUID
  org_id UUID
  token_hash TEXT
  user_agent TEXT
  ip_address TEXT
  created_at TIMESTAMP
  expires_at TIMESTAMP
  revoked_at TIMESTAMP nullable
```

---

# 5. Token Model

### Access Token (JWT)

* signed with asymmetric keys (RSA or ES256)
* private key inside `zeno-auth` only
* public JWKS exposed at `/jwks`

Claims:

```
sub: user_id
org: active_org_id
roles: ["OWNER", "ADMIN", ...]
exp, iat, iss
aud: "zenon-cloud"
```

### Refresh Token

* opaque 256-bit random string
* stored only as hash
* rotation supported
* used to issue fresh JWT

---

# 6. Secrets & Key Management

Secrets stored in **Secret Manager**:

```
zeno-auth-db-dsn  → DSN for Cloud SQL (dev)
zeno-auth-db-dsn  → DSN for Cloud SQL (prod)
```

JWT private key:

* stored **only** inside `zeno-auth` container (mounted from Secret Manager later)
* rotated manually or via automated rotation job
* public key available via `/jwks` endpoint

---

# 7. IAM & Access Boundaries

### 7.1 zeno-auth-dev permissions

* `roles/cloudsql.client`
* `roles/secretmanager.secretAccessor`
* (future) `roles/run.invoker` for internal calls

### 7.2 zeno-docs-dev permissions

* access to bucket:

    * `gs://zenon-dev-docs-raw`
    * role: `roles/storage.objectAdmin`

### 7.3 zeno-id-docs-dev permissions

* access to bucket:

    * `gs://zenon-dev-id-raw`
    * role: `roles/storage.objectAdmin`

### 7.4 Principle of Least Privilege

Every service account receives **only** the permissions required for its function.

---

# 8. Repository Structure

```
zeno-auth/
│
├─ cmd/auth/              # main entrypoint
├─ internal/
│   ├─ config/            # config loading, env & secrets
│   ├─ handler/           # HTTP endpoints
│   ├─ service/           # business logic
│   ├─ repository/        # DB queries (pgx)
│   ├─ model/             # domain models
│   └─ token/             # JWT + refresh handling
│
├─ migrations/            # SQL migrations (goose / migrate)
├─ api/proto/             # reserved for service boundary
│
├─ Dockerfile
├─ Makefile
├─ go.mod
└─ docs/
    └─ architecture.md    # this document
```

---

# 9. Environment Variables

Cloud Run injects:

```
DATABASE_URL=secret://zeno-auth-db-dsn
PORT=8080
JWT_PRIVATE_KEY=secret://zeno-auth-private-key   (future)
ENV=dev|prod
```

Local development uses `.env` instead.

---

# 10. Migration Workflow

Migrations stored in `migrations/`.

### Dev:

1. Apply automatically on `zeno-auth` startup
   or
2. Apply via GitHub Actions during deployment:

```
migrate -path=migrations -database=$DATABASE_URL up
```

### Prod:

* manual approval step
* migrations run before rolling deploy
* Cloud Run only receives traffic after successful migration

---

# 11. Build & Deploy (CI/CD)

## GitHub Actions Pipeline (.github/workflows/deploy.yml)

### Steps:

1. **Trigger:**

    * push to `main` → deploy to `dev`
    * tag `v*.*.*` → deploy to `prod`

2. **Build container:**

   ```
   docker build -t europe-west3-docker.pkg.dev/$DEV_PROJECT/zeno-auth/zeno-auth:$GIT_SHA .
   ```

3. **Push to Artifact Registry:**

   ```
   gcloud auth configure-docker
   docker push europe-west3-docker.pkg.dev/$DEV_PROJECT/zeno-auth/zeno-auth:$GIT_SHA
   ```

4. **Run migrations** (dev only, prod requires manual approval):

   ```
   migrate up
   ```

5. **Deploy to Cloud Run:**

   ```
   gcloud run deploy zeno-auth \
     --image=europe-west3-docker.pkg.dev/$DEV_PROJECT/zeno-auth/zeno-auth:$GIT_SHA \
     --platform=managed \
     --region=europe-west3 \
     --service-account=zeno-auth-dev@$DEV_PROJECT.iam.gserviceaccount.com \
     --add-cloudsql-instances=$DEV_PROJECT:europe-west3:zenon-dev-sql \
     --set-secrets=DATABASE_URL=zeno-auth-db-dsn:latest \
     --allow-unauthenticated \
     --memory=512Mi \
     --cpu=1
   ```

6. **Smoke test:**

   ```
   curl https://zeno-auth-.../health
   ```

---

# 12. API Boundary

Endpoints (MVP):

```
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
GET  /me
GET  /jwks
```

Strict separation:

* No document uploads
* No KYC processing
* No accounting logic

---

# 13. Service-to-Service Integration

### `zeno-docs` & `zeno-id-docs` consume:

* JWT tokens validated locally
* org id from claim
* roles from claim
* user id from claim

Future optional API:

```
GET /introspect
```

But not required for normal operation.

---

# 14. Security Posture

* Argon2id for password hashing
* never store refresh tokens raw
* short JWT lifetime
* long refresh lifetime with rotation
* asymmetric signing keys
* never pass secrets as environment variables inside GitHub Actions logs
* only deploy from GitHub Actions via OIDC (no store of GCP credentials)
* strict IAM boundaries
* Cloud SQL uses private socket connection (no public IP)

---

# 15. Future Extensions

* API keys for integrations
* Billing integration
* Organization invitations
* MFA support
* Passwordless login (email magic link)
* Audit log service integration
* Service Mesh (future when number of services grows)

---

# ✔ Summary

This document describes:

* what `zeno-auth` is
* what data it owns
* how tokens work
* how it integrates with the rest of ZenoN-Cloud
* how it is deployed
* how it is secured
* how CI/CD operates
* how dev/prod separation works
* how GCP resources are structured

---

Макс, брат, если хочешь — могу:

* перенести всё в Markdown с красивым форматированием (таблицы, блоки, списки)
* добавить PlantUML диаграммы
* добавить sequence diagrams для login/refresh
* собрать полноценный onboarding документ для нового инженера
