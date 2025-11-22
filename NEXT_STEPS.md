# Следующие шаги по улучшению архитектуры

## 🎯 Приоритетные задачи (High Priority)

### 1. Context Propagation & Timeouts

**Проблема:** Не все DB запросы используют context с таймаутами.

**Решение:**

```go
// В каждом методе репозитория
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()

var user model.User
err := r.db.QueryRowContext(ctx, "SELECT ...").Scan(...)
return &user, err
}
```

**Задачи:**

- [ ] Аудит всех репозиториев
- [ ] Добавить таймауты (3-5 сек для простых запросов)
- [ ] Проверить передачу ctx из handlers → services → repositories

### 2. Транзакции для критичных операций

**Проблема:** Регистрация пользователя, GDPR операции не атомарны.

**Решение:**

```go
func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*model.User, error) {
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return nil, err
}
defer tx.Rollback()

// 1. Создать пользователя
user, err := s.userRepo.CreateTx(ctx, tx, ...)
if err != nil {
return nil, err
}

// 2. Создать организацию
org, err := s.orgRepo.CreateTx(ctx, tx, ...)
if err != nil {
return nil, err
}

// 3. Создать membership
err = s.membershipRepo.CreateTx(ctx, tx, ...)
if err != nil {
return nil, err
}

return user, tx.Commit()
}
```

**Задачи:**

- [ ] Добавить методы `*Tx` в репозитории
- [ ] Обернуть регистрацию в транзакцию
- [ ] Обернуть GDPR удаление в транзакцию
- [ ] Обернуть смену пароля (удаление всех сессий) в транзакцию

### 3. Централизованный Error Mapping

**Проблема:** В handlers много switch по типам ошибок.

**Решение:**

```go
// internal/errors/mapper.go
type HTTPError struct {
StatusCode int
Code       string
Message    string
}

func MapError(err error) HTTPError {
switch {
case errors.Is(err, service.ErrInvalidCredentials):
return HTTPError{401, "invalid_credentials", "Invalid email or password"}
case errors.Is(err, service.ErrEmailExists):
return HTTPError{409, "email_exists", "Email already registered"}
case errors.Is(err, validator.ErrPasswordTooShort):
return HTTPError{400, "password_too_short", "Password must be at least 8 characters"}
default:
return HTTPError{500, "internal_error", "Internal server error"}
}
}

// В handler
func (h *AuthHandler) Login(c *gin.Context) {
// ...
accessToken, refreshToken, err := h.authService.Login(...)
if err != nil {
httpErr := errors.MapError(err)
response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
return
}
// ...
}
```

**Задачи:**

- [ ] Создать `internal/errors/mapper.go`
- [ ] Определить все domain errors
- [ ] Замапить на HTTP статусы и коды
- [ ] Рефакторить все handlers

### 4. Миграция на Unified Response

**Задачи:**

- [ ] `internal/handler/auth.go` - все endpoints
- [ ] `internal/handler/user.go` - все endpoints
- [ ] `internal/handler/consent.go` - все endpoints
- [ ] `internal/handler/gdpr.go` - все endpoints
- [ ] `internal/handler/session.go` - все endpoints
- [ ] Обновить тесты

## 📊 Medium Priority

### 5. JWT Improvements

**Добавить стандартные claims:**

```go
type Claims struct {
UserID string   `json:"user_id"`
OrgID  string   `json:"org_id"`
Roles  []string `json:"roles"`

// Стандартные claims
Issuer   string `json:"iss"` // "zeno-auth" или "https://auth.zeno.cy"
Audience string `json:"aud"` // "zeno-frontend" или список сервисов
JTI      string `json:"jti"` // Unique token ID для ревокации

jwt.RegisteredClaims
}
```

**Задачи:**

- [ ] Обновить `internal/token/jwt.go`
- [ ] Добавить валидацию `iss`, `aud` при парсинге
- [ ] Генерировать `jti` (UUID)
- [ ] Обновить тесты

### 6. JWKS Endpoint

**Цель:** Поддержка нескольких ключей и ротации.

```go
// /.well-known/jwks.json {
"keys": [
{
"kid": "2024-01",
"kty": "RSA",
"use": "sig",
"n": "...",
"e": "AQAB"
},
{
"kid": "2024-02",
"kty": "RSA",
"use": "sig",
"n": "...",
"e": "AQAB"
}
]
}
```

**Задачи:**

- [ ] Поддержка нескольких ключей в конфиге
- [ ] Добавить `kid` в JWT header
- [ ] Endpoint `/.well-known/jwks.json`
- [ ] Документация по ротации ключей

### 7. API Versioning

**Цель:** Префикс `/v1/` для всех endpoints.

```go
// Было
/auth/login
/me/profile

// Стало
/v1/auth/login
/v1/me/profile
```

**Задачи:**

- [ ] Обновить `internal/handler/router.go`
- [ ] Обновить фронтенд
- [ ] Обновить документацию
- [ ] Обновить тесты

### 8. Unit Tests

**Приоритетные тесты:**

```bash
internal/config/validator_test.go
internal/bootstrap/bootstrap_test.go
internal/middleware/request_id_test.go
internal/response/response_test.go
internal/service/auth_test.go (расширить)
internal/service/user_test.go
internal/service/password_test.go
```

**Задачи:**

- [ ] Config validator tests
- [ ] Bootstrap container tests (с моками)
- [ ] Request ID middleware tests
- [ ] Response helpers tests
- [ ] Service tests (happy path + errors)

## 🔧 Low Priority

### 9. Rate Limiting по эндпоинтам

**Разные лимиты:**

- `/auth/login` - 5 попыток / 15 мин
- `/auth/register` - 3 попытки / час
- `/auth/password/reset` - 3 попытки / час
- Остальные - 100 запросов / мин

### 10. Password Policy

**Документировать и усилить:**

- Минимум 8 символов
- Хотя бы 1 заглавная, 1 строчная, 1 цифра
- Запрет на топ-1000 паролей
- Опционально: спецсимволы

### 11. OpenAPI Documentation

**Актуализировать `api/openapi.yaml`:**

- Все endpoints
- Request/response schemas
- Error codes
- Authentication

### 12. Dev Seed Data

**Создать `scripts/seed-dev.sh`:**

```bash
# Создать тестовые данные
- 2 организации
- 5 пользователей
- Разные роли
- Тестовые сессии
```

## 📅 Roadmap

### Week 1-2: Critical Fixes

- [ ] Context timeouts (1-2 дня)
- [ ] Transactions (2-3 дня)
- [ ] Error mapping (1-2 дня)
- [ ] Unified response migration (2-3 дня)

### Week 3-4: Improvements

- [ ] JWT improvements (2 дня)
- [ ] JWKS endpoint (2 дня)
- [ ] API versioning (1 день)
- [ ] Unit tests (3-4 дня)

### Week 5+: Polish

- [ ] Rate limiting improvements
- [ ] Password policy
- [ ] OpenAPI docs
- [ ] Dev seeds
- [ ] CI improvements

## 🧪 Тестирование после изменений

```bash
# После каждого изменения
make check              # fmt + vet + lint + test
make test-integration   # интеграционные тесты
make local-test         # E2E тесты

# Перед коммитом
make cover              # проверить coverage
make local-up           # запустить локально
# Протестировать вручную основные флоу
```

## 📝 Checklist перед PR

- [ ] Код отформатирован (`make fmt`)
- [ ] Линтеры пройдены (`make lint`)
- [ ] Тесты написаны и проходят (`make test`)
- [ ] Coverage не упал (`make cover`)
- [ ] Документация обновлена
- [ ] CHANGELOG.md обновлен
- [ ] Локально протестировано (`make local-up`)

## 🎓 Полезные ресурсы

- `ARCHITECTURE_IMPROVEMENTS.md` - полный чеклист
- `IMPLEMENTATION_SUMMARY.md` - что уже сделано
- `docs/ENV_VARIABLES.md` - переменные окружения
- `Makefile` - все команды
