# ✅ Следующие шаги по улучшению архитектуры (ВЫПОЛНЕНО)

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

- [x] Аудит всех репозиториев
- [x] Добавить таймауты (3-5 сек для простых запросов)
- [x] Проверить передачу ctx из handlers → services → repositories

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

- [x] Добавить методы `*Tx` в репозитории
- [x] Обернуть регистрацию в транзакцию
- [x] Обернуть GDPR удаление в транзакцию
- [x] Обернуть смену пароля (удаление всех сессий) в транзакцию

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

- [x] Создать `internal/errors/mapper.go`
- [x] Определить все domain errors
- [x] Замапить на HTTP статусы и коды
- [x] Рефакторить все handlers

### 4. Миграция на Unified Response

**Задачи:**

- [x] `internal/handler/auth.go` - все endpoints
- [x] `internal/handler/user.go` - все endpoints
- [x] `internal/handler/consent.go` - все endpoints
- [x] `internal/handler/gdpr.go` - все endpoints
- [x] `internal/handler/session.go` - все endpoints
- [x] Обновить тесты

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

- [x] Обновить `internal/token/jwt.go`
- [x] Добавить валидацию `iss`, `aud` при парсинге
- [x] Генерировать `jti` (UUID)
- [x] Обновить тесты

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

- [x] Поддержка нескольких ключей в конфиге
- [x] Добавить `kid` в JWT header
- [x] Endpoint `/.well-known/jwks.json`
- [x] Документация по ротации ключей

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

- [x] Обновить `internal/handler/router.go`
- [x] Обновить фронтенд
- [x] Обновить документацию
- [x] Обновить тесты

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

- [x] Config validator tests
- [x] Bootstrap container tests (с моками)
- [x] Request ID middleware tests
- [x] Response helpers tests
- [x] Service tests (happy path + errors)

## 🔧 Low Priority

### 9. Rate Limiting по эндпоинтам

**Разные лимиты:**

- `/auth/login` - 5 попыток / 15 мин
- `/auth/register` - 3 попытки / час
- `/auth/password/reset` - 3 попытки / час
- Остальные - 100 запросов / мин

### 10. Password Policy ✅

**Документировать и усилить:**

- [x] Минимум 8 символов
- [x] Хотя бы 1 заглавная, 1 строчная, 1 цифра
- [x] Запрет на топ-100 паролей
- [x] Опционально: спецсимволы
- [x] Документация: `docs/PASSWORD_POLICY.md`

### 11. OpenAPI Documentation ✅

**Актуализировать `api/openapi.yaml`:**

- [x] Все endpoints с `/v1/` префиксом
- [x] Legacy endpoints помечены как deprecated
- [x] Request/response schemas
- [x] Error codes (unified response format)
- [x] Authentication (Bearer JWT)
- [x] JWKS endpoint
- [x] Password policy в описании
- [x] Версия обновлена до 1.1.0

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

### Week 1-2: Critical Fixes ✅

- [x] Context timeouts (1-2 дня)
- [x] Transactions (2-3 дня)
- [x] Error mapping (1-2 дня)
- [x] Unified response migration (2-3 дня)

### Week 3-4: Improvements ✅

- [x] JWT improvements (2 дня)
- [x] JWKS endpoint (2 дня)
- [x] API versioning (1 день)
- [x] Unit tests (3-4 дня)

### Week 5+: Polish ✅

- [x] Rate limiting improvements
- [x] Password policy
- [x] OpenAPI docs
- [x] Dev seeds
- [ ] CI improvements (optional)

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
- `docs/PASSWORD_POLICY.md` - политика паролей
- `api/openapi.yaml` - OpenAPI спецификация v1.1.0
- `Makefile` - все команды

---

## ✅ Статус выполнения

**Всего задач:** 24  
**Выполнено:** 24 (100%)  
**Осталось:** 0

### Выполненные задачи:
1. ✅ Context Propagation & Timeouts
2. ✅ Транзакции для критичных операций
3. ✅ Централизованный Error Mapping
4. ✅ Миграция на Unified Response
5. ✅ JWT Improvements
6. ✅ JWKS Endpoint
7. ✅ API Versioning
8. ✅ Unit Tests
9. ✅ Rate Limiting по эндпоинтам
10. ✅ Password Policy
11. ✅ OpenAPI Documentation
12. ✅ Dev Seed Data

**Все задачи из NEXT_STEPS.md выполнены! 🎉**
