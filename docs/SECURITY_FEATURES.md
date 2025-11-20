# 🔒 Security Features

## Реализованные функции безопасности

### 1. Rate Limiting
Защита от brute-force атак и DDoS.

**Лимиты**:
- Login: 5 попыток / 15 минут
- Register: 10 регистраций / час
- Refresh: 20 запросов / минуту
- General API: 100 запросов / минуту

**Ответ при превышении**:
```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```
HTTP Status: `429 Too Many Requests`

---

### 2. Password Validation
Строгие требования к паролям.

**Требования**:
- ✅ Минимум 8 символов
- ✅ Минимум 1 заглавная буква (A-Z)
- ✅ Минимум 1 строчная буква (a-z)
- ✅ Минимум 1 цифра (0-9)
- ✅ Не входит в список common passwords

**Примеры**:
```
❌ "password" - слишком простой
❌ "12345678" - нет букв
❌ "Password" - нет цифры
✅ "SecurePass123" - валидный
✅ "MyP@ssw0rd" - валидный
```

**Ошибки**:
```json
{
  "error": "password must be at least 8 characters long"
}
{
  "error": "password must contain at least one uppercase letter"
}
{
  "error": "password is too common, please choose a stronger password"
}
```

---

### 3. CORS Whitelist
Ограничение доступа к API только с разрешённых доменов.

**Конфигурация** (`.env`):
```env
CORS_ALLOWED_ORIGINS=http://localhost:5173,https://app.zenon-cloud.com
```

**Поведение**:
- Запросы с разрешённых доменов: ✅ Разрешены
- Запросы с других доменов: ❌ Блокируются браузером

---

### 4. Security Headers
Защита от веб-атак на уровне HTTP headers.

**Headers**:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**Защита от**:
- ✅ Clickjacking (X-Frame-Options)
- ✅ XSS атаки (X-XSS-Protection, CSP)
- ✅ MIME sniffing (X-Content-Type-Options)
- ✅ Man-in-the-middle (HSTS)

---

### 5. Input Validation & Sanitization
Валидация и очистка всех входных данных.

**Email**:
- ✅ RFC 5322 format
- ✅ Max 254 символа
- ✅ Автоматический lowercase и trim

**Full Name**:
- ✅ Max 100 символов
- ✅ Только буквы, пробелы, дефисы, апострофы
- ✅ Удаление HTML тегов
- ✅ Удаление control characters

**Примеры**:
```
Input:  "  User@Example.COM  "
Output: "user@example.com"

Input:  "John<script>alert('xss')</script>Doe"
Output: "JohnDoe"

Input:  "  John Doe  "
Output: "John Doe"
```

---

## Тестирование

### Rate Limiting
```bash
# Тест login rate limit (6 запросов - последний вернёт 429)
for i in {1..6}; do
  curl -X POST http://localhost:8080/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"wrong"}'
  echo ""
done
```

### Password Validation
```bash
# Слабый пароль
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "weak",
    "full_name": "Test User"
  }'

# Ожидаемый ответ: 400 Bad Request
# {"error":"password must be at least 8 characters long"}

# Сильный пароль
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123",
    "full_name": "Test User"
  }'

# Ожидаемый ответ: 201 Created
```

### Input Validation
```bash
# Невалидный email
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "not-an-email",
    "password": "SecurePass123",
    "full_name": "Test User"
  }'

# Ожидаемый ответ: 400 Bad Request
# {"error":"invalid email format"}

# HTML injection в имени
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123",
    "full_name": "John<script>alert(1)</script>Doe"
  }'

# HTML будет удалён автоматически
```

---

## GDPR Compliance

### Реализовано
- ✅ **Art. 32**: Appropriate technical measures (security headers, encryption)
- ✅ **Art. 5.1.f**: Integrity and confidentiality (password validation, rate limiting)

### В разработке (Week 2)
- ⏳ **Art. 15**: Right to access (data export)
- ⏳ **Art. 17**: Right to erasure (account deletion)
- ⏳ **Art. 30**: Records of processing activities (audit logs)
- ⏳ **Art. 5.1.e**: Storage limitation (data retention)

---

## Best Practices

### Для разработчиков

1. **Всегда используйте валидацию**:
```go
validator := validator.NewInputValidator()
if err := validator.ValidateEmail(email); err != nil {
    return err
}
email = validator.SanitizeEmail(email)
```

2. **Проверяйте пароли**:
```go
passwordValidator := validator.NewPasswordValidator()
if err := passwordValidator.Validate(password); err != nil {
    return err
}
```

3. **Применяйте rate limiting к новым endpoints**:
```go
auth.POST("/new-endpoint", LoginRateLimiter(), handler.NewEndpoint)
```

### Для production

1. **Настройте CORS**:
```env
CORS_ALLOWED_ORIGINS=https://app.zenon-cloud.com,https://console.zenon-cloud.com
```

2. **Используйте HTTPS**:
- Все security headers требуют HTTPS
- HSTS заставляет браузер использовать только HTTPS

3. **Мониторинг**:
- Отслеживайте 429 ответы (rate limit exceeded)
- Логируйте неудачные попытки входа
- Алерты на подозрительную активность

---

## Roadmap

### Week 2 (в разработке)
- [ ] Email verification
- [ ] Audit logging
- [ ] Data retention policy

### Week 3-4
- [ ] Password reset flow
- [ ] Change password
- [ ] Account lockout
- [ ] Session management

### Week 5+
- [ ] MFA/2FA
- [ ] Suspicious activity detection
- [ ] Data export (GDPR)
- [ ] Account deletion (GDPR)

---

## Ссылки

- [Security Implementation Plan](./security-implementation-plan.md)
- [Phase 1 Week 1 Completed](./phase1-week1-completed.md)
- [Architecture](./architecture.md)
