# ✅ zeno-auth готов к деплою на Cloud Run

## 📦 Что подготовлено

### Файлы конфигурации
- ✅ `Dockerfile` - оптимизирован для Cloud Run
- ✅ `cloudbuild.yaml` - CI/CD конфигурация
- ✅ `.gcloudignore` - исключения для деплоя
- ✅ `scripts/entrypoint.sh` - автоматические миграции при старте

### Скрипты деплоя
- ✅ `setup-gcp.sh` - настройка GCP проекта
- ✅ `deploy-cloudrun.sh` - деплой на Cloud Run
- ✅ `DEPLOY_CLOUDRUN.md` - детальная документация
- ✅ `DEPLOYMENT_CHECKLIST.md` - чеклист для деплоя

## 🚀 Быстрый старт

### 1. Настройка окружения
```bash
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="europe-west1"
```

### 2. Настройка GCP
```bash
cd /Users/maximviazov/Developer/Golang/zeno-cy/zeno-auth
./setup-gcp.sh
```

### 3. Создание секретов

**DATABASE_URL:**
```bash
echo -n "postgres://user:pass@/zeno_auth?host=/cloudsql/PROJECT:REGION:INSTANCE" | \
  gcloud secrets create zeno-auth-database-url --data-file=-
```

**JWT_PRIVATE_KEY:**
```bash
cat keys/private.pem | \
  gcloud secrets create zeno-auth-jwt-private-key --data-file=-
```

### 4. Деплой
```bash
./deploy-cloudrun.sh
```

## 🔍 Проверка

```bash
# Получить URL сервиса
SERVICE_URL=$(gcloud run services describe zeno-auth \
  --region ${GCP_REGION} \
  --format 'value(status.url)')

# Проверить health
curl ${SERVICE_URL}/health

# Проверить JWKS
curl ${SERVICE_URL}/.well-known/jwks.json
```

## 📊 Мониторинг

```bash
# Логи
gcloud run logs read zeno-auth --region ${GCP_REGION} --limit 50

# Статус
gcloud run services describe zeno-auth --region ${GCP_REGION}
```

## 🔐 Безопасность

- ✅ Секреты хранятся в Secret Manager
- ✅ Non-root пользователь в контейнере
- ✅ Миграции выполняются автоматически
- ✅ Health checks настроены
- ✅ CORS конфигурируется через env vars

## 📝 Переменные окружения

### Обязательные (через Secrets)
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_PRIVATE_KEY` - RSA private key

### Опциональные
- `ENV` - production/staging/dev (default: production)
- `PORT` - HTTP port (default: 8080)
- `LOG_LEVEL` - debug/info/warn/error (default: info)
- `LOG_FORMAT` - json/console (default: json)
- `ACCESS_TOKEN_TTL` - seconds (default: 1800)
- `REFRESH_TOKEN_TTL` - seconds (default: 1209600)
- `CORS_ALLOWED_ORIGINS` - comma-separated URLs

## 🎯 Следующие шаги

1. ✅ Подготовка завершена
2. ⏳ Создать Cloud SQL instance
3. ⏳ Создать секреты
4. ⏳ Запустить деплой
5. ⏳ Обновить frontend с новым auth URL
6. ⏳ Настроить мониторинг и алерты
