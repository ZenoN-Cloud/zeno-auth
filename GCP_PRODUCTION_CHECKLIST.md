# 🚀 GCP Production Deployment Checklist

**Project:** Zeno Auth  
**Target:** Google Cloud Run + Cloud SQL  
**Version:** 1.1.0  
**Last Updated:** 2024

---

## ✅ P0 - БЛОКЕРЫ (Must Fix Before Deploy)

### P0.1 ✅ migrate в контейнере
**Status:** ✅ FIXED  
**Location:** `Dockerfile` lines 21-23

```dockerfile
RUN wget -qO- https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate
```

**Verification:**
```bash
docker build -t test-migrate .
docker run --rm test-migrate which migrate
# Should output: /usr/local/bin/migrate
```

---

### P0.2 ⚠️ DATABASE_URL из Secret Manager

**Status:** ⚠️ NEEDS VERIFICATION  
**Required Actions:**

1. **Создать секрет в Secret Manager:**
```bash
# Формат для Cloud SQL Unix Socket:
DATABASE_URL="postgres://zeno_auth_app:STRONG_PASSWORD@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev&sslmode=disable"

# Создать секрет:
echo -n "$DATABASE_URL" | gcloud secrets create zeno-auth-database-url \
  --data-file=- \
  --replication-policy="automatic"
```

2. **Проверить секрет:**
```bash
gcloud secrets versions list zeno-auth-database-url
gcloud secrets versions access latest --secret=zeno-auth-database-url
```

3. **Убедиться, что в `gcp-deploy.sh` есть:**
```bash
--set-secrets=DATABASE_URL=zeno-auth-database-url:latest
```
✅ Уже есть в скрипте (line 127)

---

### P0.3 ⚠️ Cloud Run ↔ Cloud SQL Connection

**Status:** ⚠️ NEEDS CONFIGURATION  
**Recommended:** Unix Socket (проще для старта)

**DATABASE_URL Format:**
```
postgres://USER:PASSWORD@/DB_NAME?host=/cloudsql/INSTANCE_CONNECTION_NAME&sslmode=disable
```

**Example:**
```
postgres://zeno_auth_app:SuperSecure123!@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev&sslmode=disable
```

**Deploy Command (уже в скрипте):**
```bash
--add-cloudsql-instances="zeno-cy-dev-001:europe-west3:zeno-auth-db-dev"
```

**Alternative: Private IP + VPC Connector**
- Требует настройки VPC Connector
- DATABASE_URL: `postgres://user:pass@10.0.0.5:5432/dbname?sslmode=disable`
- Сложнее, но лучше для production

---

### P0.4 ⚠️ Service Account & IAM Roles

**Status:** ⚠️ NEEDS VERIFICATION  
**Service Account:** `zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com`

**Required Roles:**
```bash
# 1. Cloud SQL Client (обязательно)
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

# 2. Secret Manager Accessor (обязательно)
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# 3. Logging Writer (рекомендуется)
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/logging.logWriter"

# 4. Monitoring Metric Writer (опционально)
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/monitoring.metricWriter"
```

**Verification:**
```bash
gcloud projects get-iam-policy zeno-cy-dev-001 \
  --flatten="bindings[].members" \
  --filter="bindings.members:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com"
```

**Add to deploy script:**
```bash
--service-account=zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com
```

---

### P0.5 ⚠️ JWT Keys в Secret Manager

**Status:** ⚠️ NEEDS SETUP  
**Current:** Embedded public key в `internal/token/jwt_public.pem`

**Required Actions:**

1. **Создать JWT ключи (если нет):**
```bash
# Generate private key
openssl genrsa -out jwt_private.pem 2048

# Generate public key
openssl rsa -in jwt_private.pem -pubout -out jwt_public.pem

# Base64 encode для ENV (если нужно)
cat jwt_private.pem | base64 > jwt_private_base64.txt
cat jwt_public.pem | base64 > jwt_public_base64.txt
```

2. **Создать секреты:**
```bash
# Private key (ОБЯЗАТЕЛЬНО в Secret Manager)
gcloud secrets create zeno-auth-jwt-private-key \
  --data-file=jwt_private.pem \
  --replication-policy="automatic"

# Public key (опционально, можно оставить embedded)
gcloud secrets create zeno-auth-jwt-public-key \
  --data-file=jwt_public.pem \
  --replication-policy="automatic"
```

3. **Добавить в deploy:**
```bash
--set-secrets=JWT_PRIVATE_KEY=zeno-auth-jwt-private-key:latest
```

4. **⚠️ ВАЖНО: Удалить ключи из git:**
```bash
# Проверить, что нет реальных ключей:
git grep -i "BEGIN RSA PRIVATE KEY"
git grep -i "BEGIN PRIVATE KEY"

# Добавить в .gitignore:
*.pem
*.key
jwt_*
```

---

### P0.6 🔴 Debug Endpoint Security

**Status:** 🔴 CRITICAL - NEEDS FIX  
**Location:** `internal/handler/debug.go` + `internal/handler/router.go`

**Problem:** Debug endpoint защищён только `AdminAuthMiddleware()`, но маскировка слабая

**Current Code:**
```go
func maskPassword(url string) string {
    if len(url) > 50 {
        return url[:50] + "..."
    }
    return url
}
```

**Issues:**
- Показывает первые 50 символов (включая username, host)
- Доступен в production

**Fix Options:**

**Option 1: Disable in Production (RECOMMENDED)**
```go
// В router.go
if cfg.Env != "production" {
    r.GET("/debug", AdminAuthMiddleware(), Debug)
}
```

**Option 2: Improve Masking**
```go
func maskPassword(url string) string {
    if url == "" {
        return ""
    }
    // Полностью скрыть в production
    if os.Getenv("ENV") == "production" {
        return "[REDACTED]"
    }
    // В dev показать только схему
    if strings.HasPrefix(url, "postgres://") {
        return "postgres://***:***@***/***"
    }
    return "[REDACTED]"
}
```

**Action Required:** Выбрать Option 1 и отключить в production

---

## ⚠️ P1 - ВАЖНЫЕ (Should Fix Soon)

### P1.1 ⚠️ Connection Pool Limits

**Status:** ⚠️ NEEDS CONFIGURATION  
**Location:** `internal/repository/postgres/db.go`

**Current:** Использует дефолтные настройки pgxpool

**Problem:**
- Cloud SQL имеет лимиты на connections
- При масштабировании Cloud Run может исчерпать пул

**Recommended Fix:**

1. **Добавить в Config:**
```go
// internal/config/types.go
type Database struct {
    URL             string `json:"url"`
    MaxConns        int    `json:"max_conns"`
    MinConns        int    `json:"min_conns"`
    MaxConnLifetime int    `json:"max_conn_lifetime"` // seconds
}
```

2. **Применить в db.go:**
```go
func New(databaseURL string, maxConns, minConns int, maxLifetime time.Duration) (*DB, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }
    
    // Set connection pool limits
    config.MaxConns = int32(maxConns)
    config.MinConns = int32(minConns)
    config.MaxConnLifetime = maxLifetime
    config.MaxConnIdleTime = 5 * time.Minute
    
    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    // ...
}
```

3. **Recommended Values:**
```env
DB_MAX_CONNS=10          # 10 connections per instance
DB_MIN_CONNS=2           # Keep 2 warm
DB_MAX_CONN_LIFETIME=3600 # 1 hour
```

4. **Cloud Run Concurrency:**
```bash
--concurrency=80  # Default, но можно снизить до 40-50
```

**Formula:**
```
Total DB Connections = MaxConns × Max Cloud Run Instances
Example: 10 × 10 = 100 connections max
```

---

### P1.2 ⚠️ Migrations as Separate Job

**Status:** ⚠️ RECOMMENDED FOR PRODUCTION  
**Current:** Миграции запускаются в `entrypoint.sh` при старте сервиса

**Problems:**
- Деплой = миграция (нет разделения)
- При ошибке миграции сервис не стартует
- Сложнее откатывать версии

**Recommended Approach:**

**Option 1: Cloud Run Job (Recommended)**
```bash
# 1. Создать отдельный Job для миграций
gcloud run jobs create zeno-auth-migrate \
  --image=europe-west3-docker.pkg.dev/zeno-cy-dev-001/zeno-auth/zeno-auth:latest \
  --region=europe-west3 \
  --add-cloudsql-instances="zeno-cy-dev-001:europe-west3:zeno-auth-db-dev" \
  --set-secrets=DATABASE_URL=zeno-auth-database-url:latest \
  --command="/usr/local/bin/migrate" \
  --args="-path,/home/appuser/migrations,-database,$(DATABASE_URL),up"

# 2. Запускать перед деплоем
gcloud run jobs execute zeno-auth-migrate --region=europe-west3 --wait

# 3. Деплоить сервис
gcloud run deploy zeno-auth-dev ...
```

**Option 2: Separate Docker Image**
```dockerfile
# Dockerfile.migrate
FROM alpine:latest
RUN apk add --no-cache ca-certificates
RUN wget -qO- https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate
COPY migrations /migrations
ENTRYPOINT ["migrate", "-path", "/migrations", "-database"]
```

**Option 3: CI/CD Pipeline**
```yaml
# .github/workflows/deploy-prod.yml
- name: Run Migrations
  run: |
    docker run --rm \
      -e DATABASE_URL="${{ secrets.DATABASE_URL }}" \
      $IMAGE_NAME \
      /usr/local/bin/migrate -path ./migrations -database "$DATABASE_URL" up
```

**For Now:** Оставить в entrypoint, но запланировать переход на Job

---

### P1.3 ⚠️ Rate Limiting - In-Memory Store

**Status:** ⚠️ ACCEPTABLE FOR MVP  
**Location:** `internal/handler/ratelimit.go`

**Current:**
```go
store := memory.NewStore()
```

**Problem:**
- Каждый Cloud Run instance имеет свой счётчик
- Rate limit "размазан" по инстансам
- Brute-force может обойти лимиты

**Example:**
- Limit: 5 requests / 15 min
- 10 instances running
- Attacker can make: 5 × 10 = 50 requests

**Solutions:**

**Option 1: Redis Store (Recommended for Production)**
```go
import "github.com/ulule/limiter/v3/drivers/store/redis"

// В config
type RateLimit struct {
    RedisURL string `json:"redis_url"`
}

// В handler
store, err := redis.NewStore(client)
limiter := limiter.New(store, rate)
```

**Option 2: Cloud Armor (GCP Native)**
- Настроить на уровне Load Balancer
- Rate limiting по IP
- DDoS protection

**Option 3: API Gateway**
- Quota management
- Rate limiting policies

**For Now:** Оставить in-memory, но добавить в TODO

---

### P1.4 ⚠️ Request Correlation & Logging

**Status:** ⚠️ PARTIALLY IMPLEMENTED  
**Current:** Request ID middleware есть (`middleware.RequestID()`)

**Improvements Needed:**

1. **Добавить в логи:**
```go
// В middleware
log.Info().
    Str("request_id", requestID).
    Str("method", c.Request.Method).
    Str("path", c.Request.URL.Path).
    Str("ip", c.ClientIP()).
    Str("user_agent", c.Request.UserAgent()).
    Dur("latency", latency).
    Int("status", c.Writer.Status()).
    Msg("Request completed")
```

2. **Propagate Request ID:**
```go
// В каждом handler
ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
```

3. **Cloud Logging Integration:**
```go
// Добавить trace для Cloud Trace
log.Info().
    Str("logging.googleapis.com/trace", traceID).
    Str("logging.googleapis.com/spanId", spanID).
    Msg("...")
```

**Action:** Улучшить логирование после первого деплоя

---

## 📋 P2 - NICE TO HAVE (Post-MVP)

### P2.1 Encryption at Rest
- Cloud SQL: Enable automatic encryption
- Secrets: Already encrypted in Secret Manager

### P2.2 MFA/2FA
- TOTP implementation
- Backup codes

### P2.3 Email Provider
- SendGrid / AWS SES integration
- Email templates

### P2.4 Monitoring & Alerting
- Cloud Monitoring dashboards
- Alerting policies
- Error reporting

---

## 🗄️ DATABASE CHECKLIST

### Cloud SQL Setup

**1. Instance Configuration:**
```bash
# Verify instance
gcloud sql instances describe zeno-auth-db-dev

# Check status
gcloud sql instances list --filter="name:zeno-auth-db-dev"
```

**2. Database & User:**
```bash
# Connect to instance
gcloud sql connect zeno-auth-db-dev --user=postgres

# In psql:
CREATE DATABASE zeno_auth;
CREATE USER zeno_auth_app WITH PASSWORD 'STRONG_PASSWORD_HERE';
GRANT ALL PRIVILEGES ON DATABASE zeno_auth TO zeno_auth_app;

# Grant schema permissions
\c zeno_auth
GRANT ALL ON SCHEMA public TO zeno_auth_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO zeno_auth_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO zeno_auth_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO zeno_auth_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO zeno_auth_app;
```

**3. Connection String:**
```bash
# Unix Socket (Recommended)
DATABASE_URL="postgres://zeno_auth_app:PASSWORD@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev&sslmode=disable"

# Private IP (Alternative)
DATABASE_URL="postgres://zeno_auth_app:PASSWORD@10.0.0.5:5432/zeno_auth?sslmode=require"
```

**4. Create Secret:**
```bash
echo -n "$DATABASE_URL" | gcloud secrets create zeno-auth-database-url \
  --data-file=- \
  --replication-policy="automatic"
```

**5. Verify:**
```bash
gcloud secrets versions access latest --secret=zeno-auth-database-url
```

---

## 🚀 CLOUD RUN DEPLOYMENT CHECKLIST

### Pre-Deployment

- [ ] Локальные тесты пройдены: `go test ./...`
- [ ] Линтеры: `go vet ./...`, `golangci-lint run`
- [ ] Форматирование: `go fmt ./...`
- [ ] `.env.local` не в git
- [ ] Нет секретов в коде: `git grep -i "BEGIN RSA"`
- [ ] Dockerfile собирается: `docker build -t test .`

### Secrets Setup

- [ ] `zeno-auth-database-url` создан и проверен
- [ ] `zeno-auth-jwt-private-key` создан
- [ ] Service Account имеет роль `secretmanager.secretAccessor`

### IAM Setup

- [ ] Service Account создан: `zeno-auth-sa@...`
- [ ] Роль `cloudsql.client` назначена
- [ ] Роль `secretmanager.secretAccessor` назначена
- [ ] Роль `logging.logWriter` назначена

### Cloud SQL Setup

- [ ] Instance `zeno-auth-db-dev` в статусе RUNNABLE
- [ ] База `zeno_auth` создана
- [ ] Пользователь `zeno_auth_app` создан с правами
- [ ] Connection string протестирован локально

### Deployment Script

- [ ] `deploy/gcp-deploy.sh` обновлён
- [ ] `--service-account` добавлен
- [ ] `--add-cloudsql-instances` корректный
- [ ] `--set-secrets` для DATABASE_URL и JWT_PRIVATE_KEY
- [ ] `--memory=512Mi` (или больше)
- [ ] `--max-instances=10` установлен

### Post-Deployment

- [ ] `curl $SERVICE_URL/health` возвращает 200
- [ ] `curl $SERVICE_URL/health/ready` показывает `db: "up"`
- [ ] Cloud Logging не показывает ошибки подключения
- [ ] Миграции применились успешно
- [ ] `/debug` endpoint недоступен или защищён

---

## 🔧 QUICK FIXES TO APPLY NOW

### 1. Disable Debug in Production

**File:** `internal/handler/router.go`

```go
// Replace line ~86:
r.GET("/debug", AdminAuthMiddleware(), Debug)

// With:
if cfg.Env != "production" {
    r.GET("/debug", AdminAuthMiddleware(), Debug)
}
```

### 2. Add Service Account to Deploy Script

**File:** `deploy/gcp-deploy.sh`

Add after line 127:
```bash
--service-account=zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com \
```

### 3. Add JWT Secret to Deploy Script

**File:** `deploy/gcp-deploy.sh`

Add after DATABASE_URL secret:
```bash
--set-secrets=JWT_PRIVATE_KEY=zeno-auth-jwt-private-key:latest \
```

### 4. Add ENV Variables

**File:** `deploy/gcp-deploy.sh`

Add:
```bash
--set-env-vars=ENV=production,APP_NAME=zeno-auth,PORT=8080 \
```

---

## 📝 DEPLOYMENT COMMAND (Final)

```bash
gcloud run deploy zeno-auth-dev \
  --image="$IMAGE" \
  --region="$REGION" \
  --platform=managed \
  --service-account=zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com \
  --add-cloudsql-instances="$INSTANCE_CONNECTION_NAME" \
  --set-secrets=DATABASE_URL=zeno-auth-database-url:latest \
  --set-secrets=JWT_PRIVATE_KEY=zeno-auth-jwt-private-key:latest \
  --set-env-vars=ENV=production,APP_NAME=zeno-auth,PORT=8080 \
  --port=8080 \
  --memory=512Mi \
  --cpu=1 \
  --timeout=300 \
  --max-instances=10 \
  --min-instances=0 \
  --concurrency=80 \
  --allow-unauthenticated
```

---

## ✅ FINAL PRE-DEPLOY CHECKLIST

**Critical (Must Do):**
- [ ] P0.2: DATABASE_URL secret создан
- [ ] P0.3: Cloud SQL connection string корректный
- [ ] P0.4: Service Account IAM roles настроены
- [ ] P0.5: JWT keys в Secret Manager
- [ ] P0.6: Debug endpoint отключён в production

**Important (Should Do):**
- [ ] P1.1: Connection pool limits настроены
- [ ] Service account добавлен в deploy script
- [ ] ENV variables добавлены в deploy script

**Nice to Have (Can Do Later):**
- [ ] P1.2: Separate migration job
- [ ] P1.3: Redis rate limiting
- [ ] P1.4: Enhanced logging

---

## 🎯 NEXT STEPS

1. **Сейчас:** Применить Quick Fixes (5 минут)
2. **Перед деплоем:** Настроить Secrets & IAM (15 минут)
3. **Деплой:** Запустить `deploy/gcp-deploy.sh` (10 минут)
4. **После деплоя:** Проверить health checks и логи (5 минут)
5. **Post-MVP:** Реализовать P1 improvements

---

**Status:** 🟡 Ready for First Deploy (with fixes)  
**Risk Level:** Medium → Low (after P0 fixes)  
**Estimated Time to Production:** 30-45 minutes
