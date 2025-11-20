# 🔒 Zeno Auth - Security & GDPR Implementation Plan

## Критические проблемы (выявлены при аудите)

### 🚨 Высокий приоритет
1. Отсутствие валидации паролей
2. Отсутствие rate limiting (brute-force защита)
3. CORS настроен на `*` (любой origin)
4. Отсутствие email verification
5. Нет логирования критических событий
6. Отсутствие "Right to be Forgotten" (GDPR Art. 17)
7. Нет data retention policy (GDPR Art. 5.1.e)
8. Отсутствие data export (GDPR Art. 15)
9. IP адреса хранятся без явного согласия
10. Нет защиты от session hijacking
11. JWT private key в docker-compose.yml (открытый текст)
12. Отсутствие input sanitization

### 🟡 Средний приоритет
- Нет механизма смены пароля
- Нет блокировки аккаунта после N неудачных попыток
- Нет MFA (Multi-Factor Authentication)
- Нет session management UI
- Отсутствие consent management
- Нет уведомлений о подозрительной активности
- Отсутствие data breach notification mechanism
- Нет encryption at rest для sensitive fields

---

## 🔴 PHASE 1: КРИТИЧЕСКИЕ ИСПРАВЛЕНИЯ (1-2 недели)

### Week 1: Security Basics

#### 1. Rate Limiting
**Цель**: Защита от brute-force атак

**Реализация**:
- Middleware с in-memory store (golang-lru)
- Лимиты: 5 попыток входа / 15 минут на IP
- Лимиты: 10 регистраций с одного IP / час
- Лимиты: 3 попытки refresh token / минуту

**Файлы**:
- `internal/handler/ratelimit.go`
- Обновить `internal/handler/router.go`

**Библиотека**: `github.com/ulule/limiter/v3`

---

#### 2. Password Validation
**Цель**: Обеспечить сложность паролей

**Требования**:
- Минимум 8 символов
- Минимум 1 заглавная буква
- Минимум 1 строчная буква
- Минимум 1 цифра
- Опционально: 1 спецсимвол
- Проверка на common passwords (top 10k)

**Файлы**:
- `internal/validator/password.go`
- Обновить `internal/service/auth.go`

---

#### 3. CORS Whitelist
**Цель**: Ограничить доступ к API

**Реализация**:
- Заменить `*` на список разрешенных доменов
- Конфигурируемо через env: `CORS_ALLOWED_ORIGINS`
- Поддержка wildcards для dev: `*.zeno.local`

**Файлы**:
- Обновить `internal/handler/middleware.go`
- Обновить `internal/config/types.go`

---

#### 4. Security Headers
**Цель**: Защита от XSS, clickjacking, MIME sniffing

**Headers**:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
```

**Файлы**:
- Обновить `internal/handler/middleware.go`

---

#### 5. Input Validation & Sanitization
**Цель**: Защита от injection атак

**Валидация**:
- Email: RFC 5322 format
- Full name: max 100 символов, только буквы/пробелы/дефисы
- Удаление HTML/JS из текстовых полей
- Trim whitespace

**Файлы**:
- `internal/validator/input.go`
- Обновить `internal/handler/types.go` (добавить validation tags)

**Библиотека**: `github.com/go-playground/validator/v10` (уже есть)

---

### Week 2: GDPR Basics

#### 6. Email Verification
**Цель**: Подтверждение владения email (GDPR consent)

**Реализация**:
- Таблица `email_verifications` (token, user_id, expires_at)
- Endpoint `POST /auth/verify-email`
- Endpoint `POST /auth/resend-verification`
- Email отправка через SendGrid/AWS SES
- TTL токена: 24 часа
- Пользователь не может логиниться без верификации

**Файлы**:
- `migrations/005_create_email_verifications.up.sql`
- `internal/model/email_verification.go`
- `internal/repository/postgres/email_verification.go`
- `internal/service/email.go`
- `internal/handler/auth.go` (обновить)

---

#### 7. Audit Logging
**Цель**: Compliance с GDPR Art. 30, 33

**События для логирования**:
- User registered
- User logged in
- User logged out
- Login failed
- Password changed
- Email changed
- Account deleted
- Data exported
- MFA enabled/disabled

**Структура**:
```sql
audit_logs:
  id UUID
  user_id UUID (nullable)
  event_type TEXT
  event_data JSONB
  ip_address TEXT
  user_agent TEXT
  created_at TIMESTAMP
```

**Файлы**:
- `migrations/006_create_audit_logs.up.sql`
- `internal/model/audit_log.go`
- `internal/repository/postgres/audit_log.go`
- `internal/service/audit.go`
- Middleware для автоматического логирования

---

#### 8. Data Retention Policy
**Цель**: GDPR Art. 5.1.e (storage limitation)

**Политика**:
- Revoked refresh tokens: удаление через 90 дней
- Audit logs: хранение 2 года (legal requirement)
- Email verification tokens: удаление через 7 дней после истечения
- Password reset tokens: удаление через 7 дней после истечения

**Реализация**:
- Cron job (или Cloud Scheduler в GCP)
- Endpoint `POST /admin/cleanup` (для ручного запуска)

**Файлы**:
- `internal/service/cleanup.go`
- `cmd/cleanup/main.go` (отдельная команда для cron)

---

## 🟡 PHASE 2: GDPR COMPLIANCE (2-3 недели)

### Week 3: User Rights

#### 9. Right to Access (SAR - Subject Access Request)
**Цель**: GDPR Art. 15

**Endpoint**: `GET /me/data-export`

**Данные для экспорта**:
- User profile
- Organizations
- Memberships
- Active sessions
- Audit logs (последние 2 года)
- Consents

**Формат**: JSON (опционально CSV)

**Файлы**:
- `internal/service/gdpr.go`
- `internal/handler/gdpr.go`

---

#### 10. Right to be Forgotten
**Цель**: GDPR Art. 17

**Endpoint**: `DELETE /me/account`

**Процесс**:
1. Soft delete (is_deleted flag)
2. Anonymization данных:
   - email → `deleted_<uuid>@deleted.local`
   - full_name → `Deleted User`
   - password_hash → random hash
3. Удаление refresh tokens
4. Сохранение audit logs (legal requirement)
5. Cascade удаление связанных данных

**Исключения** (legal basis):
- Финансовые транзакции: 7 лет
- Audit logs: 2 года

**Файлы**:
- Обновить `internal/service/user.go`
- Обновить `internal/handler/user.go`
- `migrations/007_add_user_deleted_at.up.sql`

---

#### 11. Consent Management
**Цель**: GDPR Art. 7

**Таблица**:
```sql
user_consents:
  id UUID
  user_id UUID
  consent_type TEXT (terms, privacy, marketing, analytics)
  version TEXT
  granted BOOLEAN
  granted_at TIMESTAMP
  revoked_at TIMESTAMP
```

**Endpoints**:
- `GET /me/consents`
- `POST /me/consents`
- `DELETE /me/consents/:type`

**Файлы**:
- `migrations/008_create_user_consents.up.sql`
- `internal/model/consent.go`
- `internal/repository/postgres/consent.go`
- `internal/service/consent.go`
- `internal/handler/consent.go`

---

### Week 4: Advanced Security

#### 12. Password Reset Flow
**Endpoint**: 
- `POST /auth/forgot-password` (отправка email)
- `POST /auth/reset-password` (сброс с токеном)

**Таблица**:
```sql
password_reset_tokens:
  id UUID
  user_id UUID
  token_hash TEXT
  expires_at TIMESTAMP
  used_at TIMESTAMP
```

**TTL**: 15 минут

**Файлы**:
- `migrations/009_create_password_reset_tokens.up.sql`
- `internal/model/password_reset.go`
- `internal/repository/postgres/password_reset.go`
- `internal/service/password_reset.go`
- `internal/handler/auth.go` (обновить)

---

#### 13. Change Password
**Endpoint**: `POST /me/change-password`

**Требования**:
- Текущий пароль обязателен
- Новый пароль проходит валидацию
- Отзыв всех refresh tokens (force re-login)
- Email уведомление
- Audit log

**Файлы**:
- Обновить `internal/service/user.go`
- Обновить `internal/handler/user.go`

---

#### 14. Account Lockout
**Цель**: Защита от brute-force

**Логика**:
- После 5 неудачных попыток → блокировка на 30 минут
- Email уведомление о блокировке
- Endpoint для разблокировки (admin или email link)

**Таблица**:
```sql
ALTER TABLE users ADD COLUMN failed_login_attempts INT DEFAULT 0;
ALTER TABLE users ADD COLUMN locked_until TIMESTAMP;
```

**Файлы**:
- `migrations/010_add_user_lockout.up.sql`
- Обновить `internal/service/auth.go`

---

### Week 5: Session Management

#### 15. Session Fingerprinting
**Цель**: Защита от session hijacking

**Fingerprint**:
- User-Agent hash
- IP address (первые 3 октета)
- Accept-Language
- Опционально: TLS fingerprint

**Реализация**:
- Добавить `fingerprint_hash` в `refresh_tokens`
- Проверка при refresh

**Файлы**:
- `migrations/011_add_fingerprint_to_refresh_tokens.up.sql`
- `internal/token/fingerprint.go`
- Обновить `internal/service/auth.go`

---

#### 16. Active Sessions Management
**Endpoints**:
- `GET /me/sessions` - список активных сессий
- `DELETE /me/sessions/:id` - отзыв конкретной сессии
- `DELETE /me/sessions` - отзыв всех кроме текущей

**Данные сессии**:
- Device info (parsed User-Agent)
- Location (IP → GeoIP)
- Last activity
- Current session indicator

**Файлы**:
- `internal/service/session.go`
- `internal/handler/session.go`

---

## 🟢 PHASE 3: PRODUCTION READINESS (2-3 недели)

### Week 6: Monitoring

#### 17. Structured Logging
- Correlation IDs для трейсинга
- Structured fields (user_id, org_id, ip, etc.)
- Log levels по окружению

#### 18. Metrics (Prometheus)
**Metrics**:
- `auth_registrations_total`
- `auth_logins_total`
- `auth_login_failures_total`
- `auth_token_refreshes_total`
- `auth_request_duration_seconds`
- `auth_active_sessions`

**Endpoint**: `GET /metrics`

#### 19. Enhanced Health Checks
- DB connection check
- Redis check (если используется)
- Disk space check
- Memory check

---

### Week 7: Advanced Features

#### 20. MFA/2FA (TOTP)
**Таблица**:
```sql
mfa_secrets:
  id UUID
  user_id UUID
  secret TEXT (encrypted)
  backup_codes TEXT[] (encrypted)
  enabled_at TIMESTAMP
```

**Endpoints**:
- `POST /me/mfa/enable` (генерация QR кода)
- `POST /me/mfa/verify` (подтверждение кода)
- `POST /me/mfa/disable`
- `GET /me/mfa/backup-codes`

**Библиотека**: `github.com/pquerna/otp`

---

#### 21. Email Notifications
**События**:
- Новый вход с нового устройства
- Смена пароля
- Смена email
- Подозрительная активность
- Account lockout

**Шаблоны**: HTML + plain text

---

### Week 8: Organization Features

#### 22. Organization Invitations
**Таблица**:
```sql
org_invitations:
  id UUID
  org_id UUID
  email TEXT
  role TEXT
  invited_by UUID
  token_hash TEXT
  expires_at TIMESTAMP
  accepted_at TIMESTAMP
```

**Endpoints**:
- `POST /orgs/:id/invite`
- `GET /invitations/:token`
- `POST /invitations/:token/accept`
- `DELETE /invitations/:id`

---

#### 23. Role Management
**Endpoints**:
- `PATCH /orgs/:id/members/:user_id/role`
- `DELETE /orgs/:id/members/:user_id`

**Permission checks**: Только OWNER/ADMIN

---

## 🔵 PHASE 4: ADVANCED COMPLIANCE (2 недели)

### Week 9: Data Protection

#### 24. Encryption at Rest
- Шифрование `full_name`, `email` (опционально)
- AES-256-GCM
- Key management через GCP KMS
- Transparent для приложения

#### 25. Data Breach Detection
- Мониторинг аномальной активности
- Алерты в Slack/PagerDuty
- Автоматический lockout при подозрении

#### 26. Compliance Reports
**Endpoint**: `GET /admin/compliance/report`

**Данные**:
- Количество SAR requests
- Количество deletion requests
- Среднее время ответа
- Audit trail summary

---

### Week 10: Final Touches

#### 27. API Versioning
- Переход на `/v1/auth/...`
- Deprecation headers
- Backward compatibility

#### 28. Documentation
- OpenAPI/Swagger spec
- Privacy Policy template
- Terms of Service template
- GDPR compliance documentation
- DPA (Data Processing Agreement) template

#### 29. Security Audit
- OWASP Top 10 check
- Penetration testing
- Dependency vulnerability scan
- Code review

---

## Оценка трудозатрат

| Phase | Недели | Человеко-дни |
|-------|--------|--------------|
| Phase 1 | 2 | 10 |
| Phase 2 | 3 | 15 |
| Phase 3 | 3 | 15 |
| Phase 4 | 2 | 10 |
| **ИТОГО** | **10** | **50** |

---

## MVP Приоритизация (для быстрого запуска)

### Must Have (2 недели):
1. ✅ Rate limiting
2. ✅ Password validation
3. ✅ CORS whitelist
4. ✅ Email verification
5. ✅ Audit logging
6. ✅ Data export (SAR)
7. ✅ Account deletion
8. ✅ Password reset

### Should Have (1 неделя):
9. Session management
10. Security headers
11. MFA

### Nice to Have (после MVP):
12. Advanced monitoring
13. Encryption at rest
14. Organization invitations

---

## Технический стек

```go
// Rate Limiting
github.com/ulule/limiter/v3

// Email
github.com/sendgrid/sendgrid-go
// или AWS SES SDK

// Metrics
github.com/prometheus/client_golang

// Tracing
go.opentelemetry.io/otel

// MFA
github.com/pquerna/otp

// Encryption
golang.org/x/crypto/nacl/secretbox

// Redis (опционально)
github.com/redis/go-redis/v9

// Validation (уже есть)
github.com/go-playground/validator/v10
```

---

## Compliance Checklist

### GDPR Requirements
- [ ] Right to access (Art. 15)
- [ ] Right to rectification (Art. 16)
- [ ] Right to erasure (Art. 17)
- [ ] Right to data portability (Art. 20)
- [ ] Consent management (Art. 7)
- [ ] Data retention policies (Art. 5.1.e)
- [ ] Breach notification (Art. 33)
- [ ] Privacy by design (Art. 25)
- [ ] Data protection impact assessment (Art. 35)

### Security Best Practices
- [ ] Password hashing (Argon2id) ✅
- [ ] Rate limiting
- [ ] Input validation
- [ ] Output encoding
- [ ] HTTPS only
- [ ] Secure headers
- [ ] CSRF protection
- [ ] SQL injection prevention ✅ (pgx)
- [ ] XSS prevention
- [ ] Session management
- [ ] MFA support
- [ ] Audit logging

---

**Статус**: Ready for Phase 1 implementation
**Дата создания**: 2024
**Автор**: Security Audit Team
