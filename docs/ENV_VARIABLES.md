# Environment Variables

Полная документация переменных окружения для Zeno Auth.

## Обязательные переменные

### Database

- **`DATABASE_URL`** (обязательно)
    - Формат: `postgres://user:password@host:port/dbname?sslmode=disable`
    - Пример: `postgres://zeno_auth:secret@localhost:5432/zeno_auth?sslmode=disable`
    - Описание: URL подключения к PostgreSQL

### JWT Keys

- **`JWT_PRIVATE_KEY`** (обязательно в production)
    - Формат: Base64-encoded RSA private key
    - Описание: Приватный ключ для подписи JWT токенов
    - Dev: можно использовать файл `jwt-private.pem`

- **`JWT_PUBLIC_KEY`** (обязательно в production)
    - Формат: Base64-encoded RSA public key
    - Описание: Публичный ключ для верификации JWT токенов
    - Dev: можно использовать файл `jwt-public.pem`

## Опциональные переменные

### Server

- **`SERVER_PORT`** (по умолчанию: `8080`)
    - Описание: Порт HTTP сервера

- **`CORS_ALLOWED_ORIGINS`** (по умолчанию: `http://localhost:5173`)
    - Формат: Список через запятую
    - Пример: `https://app.zeno.cy,https://admin.zeno.cy`
    - Описание: Разрешенные CORS origins
    - ⚠️ В production НЕ использовать `*`

### JWT Configuration

- **`JWT_ACCESS_TOKEN_TTL`** (по умолчанию: `900` = 15 минут)
    - Формат: Секунды
    - Диапазон: 300-86400 (5 мин - 24 часа)
    - Описание: Время жизни access token

- **`JWT_REFRESH_TOKEN_TTL`** (по умолчанию: `604800` = 7 дней)
    - Формат: Секунды
    - Диапазон: 86400-2592000 (1-30 дней)
    - Описание: Время жизни refresh token

### Application

- **`ENV`** (по умолчанию: `development`)
    - Значения: `development`, `staging`, `production`
    - Описание: Окружение приложения

- **`APP_NAME`** (по умолчанию: `zeno-auth`)
    - Описание: Название приложения для логов

- **`TIMEZONE`** (по умолчанию: `UTC`)
    - Пример: `Europe/Moscow`, `America/New_York`
    - Описание: Временная зона приложения

### Logging

- **`LOG_LEVEL`** (по умолчанию: `info`)
    - Значения: `debug`, `info`, `warn`, `error`
    - Описание: Уровень логирования

- **`LOG_FORMAT`** (по умолчанию: `json`)
    - Значения: `json`, `console`
    - Описание: Формат логов

- **`LOG_FILE`** (опционально)
    - Пример: `/var/log/zeno-auth.log`
    - Описание: Путь к файлу логов (если не задан - только stdout)

## Production секреты

В production окружении **ОБЯЗАТЕЛЬНО** использовать Secret Manager:

### Google Cloud Secret Manager

```bash
# Создание секретов
gcloud secrets create zeno-auth-jwt-private --data-file=jwt-private.pem
gcloud secrets create zeno-auth-jwt-public --data-file=jwt-public.pem
gcloud secrets create zeno-auth-db-url --data-file=-

# Использование в Cloud Run
gcloud run deploy zeno-auth \
  --set-secrets="JWT_PRIVATE_KEY=zeno-auth-jwt-private:latest" \
  --set-secrets="JWT_PUBLIC_KEY=zeno-auth-jwt-public:latest" \
  --set-secrets="DATABASE_URL=zeno-auth-db-url:latest"
```

## Примеры конфигураций

### Development (.env.local)

```env
ENV=development
DATABASE_URL=postgres://zeno_auth:zeno_auth@localhost:5432/zeno_auth?sslmode=disable
JWT_PRIVATE_KEY=<base64-encoded-key>
JWT_PUBLIC_KEY=<base64-encoded-key>
SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
LOG_LEVEL=debug
LOG_FORMAT=console
```

### Production

```env
ENV=production
DATABASE_URL=<from-secret-manager>
JWT_PRIVATE_KEY=<from-secret-manager>
JWT_PUBLIC_KEY=<from-secret-manager>
SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=https://app.zeno.cy,https://admin.zeno.cy
LOG_LEVEL=info
LOG_FORMAT=json
JWT_ACCESS_TOKEN_TTL=900
JWT_REFRESH_TOKEN_TTL=604800
```

## Валидация

При старте приложения все переменные проходят валидацию:

- Проверка обязательных полей
- Проверка диапазонов значений
- Проверка формата (email, URL, timezone)
- Проверка безопасности (CORS wildcard в production)

Если валидация не прошла - приложение не запустится с подробным описанием ошибок.

## Генерация JWT ключей

```bash
# Генерация ключей
make generate-keys

# Или вручную
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

# Конвертация в base64 для ENV
cat jwt-private.pem | base64
cat jwt-public.pem | base64
```
