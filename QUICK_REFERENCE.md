# 🚀 Quick Reference - Zeno Auth

Быстрая справка по новым возможностям и командам.

## 📦 Новые пакеты

### Bootstrap Container

```go
import "github.com/ZenoN-Cloud/zeno-auth/internal/bootstrap"

// Создание контейнера со всеми зависимостями
container, err := bootstrap.BuildContainer(cfg)

// Использование
container.AuthService.Login(ctx, email, password, userAgent, ip)
container.UserService.GetByID(ctx, userID)
```

### Request ID Middleware

```go
import "github.com/ZenoN-Cloud/zeno-auth/internal/middleware"

// В router
r.Use(middleware.RequestID())

// Получение request ID
requestID := middleware.GetRequestID(ctx)

// Логирование с request ID
log.Ctx(ctx).Info().Msg("Processing request") // автоматически добавит request_id
```

### Unified Response

```go
import "github.com/ZenoN-Cloud/zeno-auth/internal/response"

// Success
response.Success(c, http.StatusOK, gin.H{"user": user})
// {"status": "ok", "data": {"user": {...}}}

// Errors
response.BadRequest(c, "Invalid email format")
response.Unauthorized(c, "Invalid credentials")
response.Forbidden(c, "Access denied")
response.NotFound(c, "User not found")
response.Conflict(c, "Email already exists")
response.InternalError(c, "Something went wrong")
response.ServiceUnavailable(c, "Service temporarily unavailable")

// Custom error
response.Error(c, http.StatusTeapot, "im_a_teapot", "I'm a teapot")
```

### Config Validation

```go
import "github.com/ZenoN-Cloud/zeno-auth/internal/config"

cfg, err := config.Load()
if err != nil {
    return err
}

// Валидация (автоматически вызывается в bootstrap)
if err := cfg.Validate(); err != nil {
    log.Fatal().Err(err).Msg("Invalid configuration")
}
```

## 🛠️ Make команды

### Разработка

```bash
make dev              # Запуск с .env
make run              # Запуск без .env
make build            # Сборка бинарника
```

### Локальное окружение

```bash
make local-up         # Запустить все сервисы (Docker Compose)
make local-down       # Остановить сервисы
make local-restart    # Перезапустить
make local-rebuild    # Пересобрать и запустить
make local-logs       # Показать логи
make local-logs-auth  # Логи только auth сервиса
make local-status     # Статус сервисов
make local-clean      # Удалить все данные
make local-test       # E2E тесты
```

### Проверки кода

```bash
make fmt              # go fmt
make fmt-strict       # gofumpt (строгое форматирование)
make vet              # go vet
make lint             # все линтеры
make staticcheck      # статический анализ
make check            # fmt + vet + lint + test (все проверки)
```

### Тесты

```bash
make test             # Все тесты
make test-unit        # Только unit тесты
make test-integration # Интеграционные тесты (Docker)
make test-e2e         # E2E тесты
make cover            # Coverage с HTML отчетом
```

### Миграции

```bash
make migrate-up       # Применить миграции
make migrate-down     # Откатить миграции
make migrate-reset    # down + up (сброс БД)
make migrate-create NAME=add_users_table  # Создать новую миграцию
```

### Зависимости

```bash
make deps             # Установить зависимости
make install-tools    # Установить dev tools (golangci-lint, gofumpt, staticcheck)
```

### Утилиты

```bash
make generate-keys    # Сгенерировать JWT ключи
make gen-key          # Алиас для generate-keys
make clean            # Удалить бинарники
```

## 🔧 Переменные окружения

### Обязательные

```env
DATABASE_URL=postgres://user:pass@host:5432/dbname
JWT_PRIVATE_KEY=<base64-encoded-key>
JWT_PUBLIC_KEY=<base64-encoded-key>
```

### Опциональные

```env
SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
ENV=development|staging|production
LOG_LEVEL=debug|info|warn|error
LOG_FORMAT=json|console
JWT_ACCESS_TOKEN_TTL=900
JWT_REFRESH_TOKEN_TTL=604800
```

Подробнее: `docs/ENV_VARIABLES.md`

## 📝 Примеры кода

### Handler с unified response

```go
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request format")
        return
    }

    ctx := c.Request.Context()
    accessToken, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password, userAgent, ip)
    if err != nil {
        if errors.Is(err, service.ErrInvalidCredentials) {
            response.Unauthorized(c, "Invalid credentials")
            return
        }
        response.InternalError(c, "Login failed")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "access_token": accessToken,
        "refresh_token": refreshToken,
    })
}
```

### Service с context

```go
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ip string) (string, string, error) {
    // Получить request ID для логов
    requestID := middleware.GetRequestID(ctx)
    
    // Логирование с context (автоматически добавит request_id)
    log.Ctx(ctx).Info().
        Str("email", email).
        Str("ip", ip).
        Msg("Login attempt")

    // Вызов репозитория с context
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return "", "", err
    }

    // ...
}
```

### Repository с timeout

```go
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    // Добавить таймаут для DB запроса
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    var user model.User
    err := r.db.QueryRowContext(ctx, 
        "SELECT id, email, full_name FROM users WHERE email = $1",
        email,
    ).Scan(&user.ID, &user.Email, &user.FullName)
    
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    return &user, err
}
```

### Тест с mock container

```go
func TestAuthHandler_Login(t *testing.T) {
    // Mock сервисы
    mockAuth := &MockAuthService{}
    mockAuth.On("Login", mock.Anything, "test@example.com", "password", mock.Anything, mock.Anything).
        Return("access_token", "refresh_token", nil)

    // Создать handler с mock
    handler := NewAuthHandler(mockAuth, nil, nil, nil, nil)

    // Тест
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/auth/login", body)
    
    handler.Login(c)

    assert.Equal(t, http.StatusOK, w.Code)
    // ...
}
```

## 🐛 Отладка

### Логи с request ID

```bash
# Все логи содержат request_id
{"level":"info","request_id":"550e8400-e29b-41d4-a716-446655440000","msg":"User logged in"}

# Фильтрация по request_id
docker logs zeno-auth | grep "550e8400-e29b-41d4-a716-446655440000"
```

### Проверка конфигурации

```bash
# Приложение не запустится если конфиг невалиден
go run cmd/auth/main.go
# Вывод:
# Config validation failed:
#   - DATABASE_URL is required
#   - JWT_PRIVATE_KEY is required in production
```

### Health checks

```bash
# Basic health
curl http://localhost:8080/health

# Readiness (с проверкой БД)
curl http://localhost:8080/health/ready

# Liveness
curl http://localhost:8080/health/live
```

## 📚 Документация

### Основные документы

- `README.md` - Общее описание
- `QUICKSTART.md` - Быстрый старт
- `REFACTORING_COMPLETE.md` - ⭐ Что реализовано

### Архитектура

- `ARCHITECTURE_IMPROVEMENTS.md` - Чеклист (40 задач)
- `IMPLEMENTATION_SUMMARY.md` - Детальное описание
- `NEXT_STEPS.md` - ⭐ Следующие шаги

### Операции

- `docs/ENV_VARIABLES.md` - ⭐ Переменные окружения
- `docs/architecture.md` - Архитектура
- `SECURITY_CHECKLIST.md` - Security checklist

## 🎯 Workflow

### Новая фича

```bash
# 1. Создать ветку
git checkout -b feature/new-feature

# 2. Разработка
make dev  # запустить локально

# 3. Проверки
make check  # fmt + vet + lint + test
make cover  # проверить coverage

# 4. Коммит
git add .
git commit -m "feat: add new feature"

# 5. Push
git push origin feature/new-feature
```

### Исправление бага

```bash
# 1. Воспроизвести
make local-up
make local-logs

# 2. Написать тест
# test/bug_test.go

# 3. Исправить
# internal/service/...

# 4. Проверить
make test
make local-test

# 5. Коммит
git commit -m "fix: resolve issue with ..."
```

## 🚨 Troubleshooting

### Проблема: Приложение не запускается

```bash
# Проверить конфигурацию
cat .env.local

# Проверить БД
docker ps | grep postgres
docker logs zeno-auth-postgres

# Проверить миграции
make migrate-up
```

### Проблема: Тесты падают

```bash
# Запустить с verbose
go test -v ./...

# Запустить конкретный тест
go test -v -run TestAuthService_Login ./internal/service

# Проверить coverage
make cover
```

### Проблема: Не компилируется

```bash
# Обновить зависимости
make deps

# Проверить синтаксис
make vet

# Форматирование
make fmt
```

---

**Быстрая помощь:**

- Проблемы с конфигом → `docs/ENV_VARIABLES.md`
- Проблемы с архитектурой → `IMPLEMENTATION_SUMMARY.md`
- Следующие задачи → `NEXT_STEPS.md`
- Все команды → `make help` или этот файл
