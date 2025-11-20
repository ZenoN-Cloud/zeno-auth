# 🚀 Локальная разработка Zeno Auth

Руководство по запуску и обкатке **Zeno Auth** локально через Docker Compose.

## 📋 Предварительные требования

- Docker Desktop (или Docker Engine + Docker Compose)
- Make (опционально, для удобства)
- Git

## 🏃 Быстрый старт

### 1. Запустить все сервисы

```bash
make local-up
```

Или без Make:

```bash
docker-compose up -d
```

Это запустит:
- **PostgreSQL** на порту `5432`
- **Zeno Auth API** на порту `8080`
- **pgAdmin** на порту `5050`

### 2. Проверить статус

```bash
make local-status
```

Или:

```bash
docker-compose ps
```

### 3. Проверить работу API

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🔧 Основные команды

| Команда | Описание |
|---------|----------|
| `make local-up` | Запустить все сервисы |
| `make local-down` | Остановить все сервисы |
| `make local-clean` | Остановить и удалить все данные (volumes) |
| `make local-logs` | Показать логи всех сервисов |
| `make local-logs-auth` | Показать логи только auth сервиса |
| `make local-restart` | Перезапустить сервисы |
| `make local-rebuild` | Пересобрать и запустить |
| `make local-status` | Статус сервисов |

## 🌐 Доступ к сервисам

### Zeno Console (Frontend)

- **URL:** `http://localhost:5173`
- **Логин/Регистрация:** Доступны сразу

### API Endpoints

- **Base URL:** `http://localhost:8080`
- **Health Check:** `http://localhost:8080/health`
- **JWKS:** `http://localhost:8080/jwks`

### pgAdmin (Web UI для PostgreSQL)

- **URL:** `http://localhost:5050`
- **Email:** `admin@zeno.local`
- **Password:** `admin`

#### Подключение к БД в pgAdmin:

1. Открой `http://localhost:5050`
2. Войди с учетными данными выше
3. Добавь новый сервер:
   - **Name:** `Zeno Auth Local`
   - **Host:** `postgres` (имя контейнера)
   - **Port:** `5432`
   - **Database:** `zeno_auth`
   - **Username:** `zeno_auth`
   - **Password:** `devpassword`

### PostgreSQL (прямое подключение)

Если хочешь подключиться напрямую с хоста:

```bash
psql postgres://zeno_auth:devpassword@localhost:5432/zeno_auth
```

Или через любой SQL клиент:
- **Host:** `localhost`
- **Port:** `5432`
- **Database:** `zeno_auth`
- **User:** `zeno_auth`
- **Password:** `devpassword`

## 🧪 Тестирование API

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "full_name": "Test User"
  }'
```

### Логин

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

### Получить текущего пользователя

```bash
# Сначала получи access_token из логина
ACCESS_TOKEN="your_access_token_here"

curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Создать организацию

```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "slug": "my-company"
  }'
```

## 🔍 Отладка

### Просмотр логов

Все сервисы:
```bash
docker-compose logs -f
```

Только auth:
```bash
docker-compose logs -f zeno-auth
```

Только postgres:
```bash
docker-compose logs -f postgres
```

### Зайти в контейнер

```bash
# Auth сервис
docker exec -it zeno-auth-app sh

# PostgreSQL
docker exec -it zeno-auth-postgres psql -U zeno_auth -d zeno_auth
```

### Проверить миграции

```bash
docker exec -it zeno-auth-postgres psql -U zeno_auth -d zeno_auth -c "\dt"
```

Должны быть таблицы:
- `users`
- `organizations`
- `org_memberships`
- `refresh_tokens`
- `schema_migrations`

## 🛠️ Разработка

### Изменение кода

После изменения кода нужно пересобрать образ:

```bash
make local-rebuild
```

Или:

```bash
docker-compose up -d --build
```

### Сброс базы данных

Если нужно начать с чистой БД:

```bash
make local-clean
make local-up
```

### Переменные окружения

Все переменные настроены в `docker-compose.yml`. Для изменений:

1. Отредактируй `docker-compose.yml`
2. Перезапусти: `make local-restart`

## 📊 Мониторинг

### Проверка здоровья сервисов

```bash
# Health check API
curl http://localhost:8080/health

# Проверка БД
docker exec zeno-auth-postgres pg_isready -U zeno_auth
```

### Статистика контейнеров

```bash
docker stats zeno-auth-app zeno-auth-postgres
```

## 🐛 Решение проблем

### Порты заняты

Если порты `5432`, `8080` или `5050` заняты, измени их в `docker-compose.yml`:

```yaml
ports:
  - "5433:5432"  # Вместо 5432:5432
```

### Контейнер не запускается

Проверь логи:
```bash
docker-compose logs zeno-auth
```

### База данных не готова

Подожди пока пройдет healthcheck:
```bash
docker-compose ps
```

Статус должен быть `healthy`.

### Очистка всего

Полная очистка Docker:
```bash
make local-clean
docker system prune -a --volumes
```

## 🔐 Безопасность

⚠️ **ВАЖНО:** Ключи и пароли в `docker-compose.yml` и `.env.local` предназначены ТОЛЬКО для локальной разработки!

**НИКОГДА** не используй их в production!

## 📚 Дополнительные ресурсы

- [README.md](./README.md) - Основная документация
- [docs/architecture.md](./docs/architecture.md) - Архитектура
- [docs/implementation-plan.md](./docs/implementation-plan.md) - План реализации
- [deploy/README.md](./deploy/README.md) - Деплой в production

## 🤝 Следующие шаги

После успешной обкатки локально:

1. ✅ Протестируй все API endpoints
2. ✅ Проверь работу с организациями и ролями
3. ✅ Убедись что JWT токены работают корректно
4. ✅ Проверь refresh token flow
5. 🚀 Переходи к деплою в dev окружение

---

**Удачи с разработкой! 🎉**
