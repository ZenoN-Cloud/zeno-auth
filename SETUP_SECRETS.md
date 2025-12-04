# 🔐 Создание секретов в GCP (EU-compliant)

> **⚠️ Предупреждение безопасности:** Никогда не используйте простые пароли в production. Всегда генерируйте случайные пароли и храните секреты безопасно.

## Шаг 1: Авторизация

```bash
# Авторизуйся в gcloud
gcloud auth login

# Установи проект
gcloud config set project zeno-cy-dev-001
gcloud config set compute/region europe-west3
```

## Шаг 2: Запуск скрипта

```bash
# Перейди в директорию проекта
cd /path/to/your/zeno-auth
./create-secrets.sh
```

## Что будет создано:

### 1. JWT_PRIVATE_KEY ✅
- Источник: `keys/private.pem`
- Регионы: `europe-west3, europe-west1` (только EU)
- Автоматически

### 2. JWT_PUBLIC_KEY ✅
- Источник: `keys/public.pem`
- Регионы: `europe-west3, europe-west1` (только EU)
- Автоматически

### 3. DATABASE_URL ⚠️
- Нужно ввести вручную
- Формат: `postgres://USER:PASSWORD@/DB_NAME?host=/cloudsql/INSTANCE_CONNECTION_NAME`
- Пример: `postgres://${DB_USER}:${SECURE_PASSWORD}@/${DB_NAME}?host=/cloudsql/${PROJECT_ID}:${REGION}:${INSTANCE_NAME}`
- Реальный пример: `postgres://zeno_auth:$(openssl rand -base64 32)@/zeno_auth?host=/cloudsql/zeno-cy-dev-001:europe-west3:zeno-auth-db-dev`

### 4. SENDGRID_API_KEY (опционально)
- Для email уведомлений
- Можно пропустить (Enter)

## Проверка после создания:

```bash
# Список всех секретов
gcloud secrets list --filter="name:zeno-auth"

# Проверка конкретного секрета
gcloud secrets describe zeno-auth-jwt-private-key

# Проверка репликации (должно быть только EU)
gcloud secrets describe zeno-auth-jwt-private-key --format="value(replication)"
```

## Если нужно пересоздать секрет:

```bash
# Удалить
gcloud secrets delete zeno-auth-jwt-private-key

# Создать заново
./create-secrets.sh
```

## 🔒 Рекомендации по безопасности

### Генерация безопасных паролей:
```bash
# Генерация случайного пароля (32 символа)
DB_PASSWORD=$(openssl rand -base64 32)
echo "Generated password: $DB_PASSWORD"

# Или с помощью pwgen
pwgen -s 32 1
```

### Проверка силы пароля:
- Минимум 16 символов для production
- Смешанные символы: буквы, цифры, спецсимволы
- Никаких словарных слов

## EU Compliance ✅

Все секреты хранятся **только в EU регионах**:
- 🇩🇪 europe-west3 (Frankfurt, Germany)
- 🇧🇪 europe-west1 (Belgium)

Соответствует GDPR Article 44-50 (Data transfers).
