#!/bin/bash
set -e

PROJECT_ID="zeno-cy-dev-001"
REGION="europe-west3"

echo "🔐 Создание секретов для zeno-auth в GCP"
echo "Project: $PROJECT_ID"
echo ""

# Установить проект
gcloud config set project "$PROJECT_ID"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  JWT_PRIVATE_KEY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if gcloud secrets describe zeno-auth-jwt-private-key &> /dev/null; then
    echo "✅ Секрет уже существует"
    gcloud secrets versions list zeno-auth-jwt-private-key --limit=1
else
    if [ ! -f "keys/private.pem" ]; then
        echo "❌ Файл keys/private.pem не найден"
        echo "Создайте ключи командой: make generate-keys"
        exit 1
    fi
    echo "📝 Создаю секрет из keys/private.pem (EU-only)..."
    if gcloud secrets create zeno-auth-jwt-private-key \
        --data-file=keys/private.pem \
        --replication-policy="user-managed" \
        --locations="europe-west3,europe-west1" 2>/dev/null; then
        echo "✅ Создан (только EU регионы)"
    else
        echo "❌ Ошибка создания секрета"
        exit 1
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  JWT_PUBLIC_KEY (опционально)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if gcloud secrets describe zeno-auth-jwt-public-key &> /dev/null; then
    echo "✅ Секрет уже существует"
else
    echo "📝 Создаю секрет из keys/public.pem (EU-only)..."
    gcloud secrets create zeno-auth-jwt-public-key \
        --data-file=keys/public.pem \
        --replication-policy="user-managed" \
        --locations="europe-west3,europe-west1"
    echo "✅ Создан (только EU регионы)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  DATABASE_URL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if gcloud secrets describe zeno-auth-database-url &> /dev/null; then
    echo "✅ Секрет уже существует"
    gcloud secrets versions list zeno-auth-database-url --limit=1
else
    echo "❌ Секрет не найден"
    echo ""
    echo "Формат DATABASE_URL для Cloud SQL:"
    echo "postgres://USER:PASSWORD@/DB_NAME?host=/cloudsql/INSTANCE_CONNECTION_NAME"
    echo ""
    echo "Пример:"
    echo "postgres://zeno_auth:MyPass123@/zeno_auth?host=/cloudsql/$PROJECT_ID:$REGION:zeno-auth-db-dev"
    echo ""
    read -p "Введи DATABASE_URL: " DATABASE_URL
    
    if [ -n "$DATABASE_URL" ]; then
        echo -n "$DATABASE_URL" | gcloud secrets create zeno-auth-database-url \
            --data-file=- \
            --replication-policy="user-managed" \
            --locations="europe-west3,europe-west1"
        echo "✅ Создан (только EU регионы)"
    else
        echo "⚠️  Пропущено"
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  SENDGRID_API_KEY (опционально)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if gcloud secrets describe zeno-auth-sendgrid-api-key &> /dev/null; then
    echo "✅ Секрет уже существует"
else
    echo "❌ Секрет не найден"
    read -p "Введи SendGrid API Key (или Enter для пропуска): " SENDGRID_KEY
    
    if [ -n "$SENDGRID_KEY" ]; then
        echo -n "$SENDGRID_KEY" | gcloud secrets create zeno-auth-sendgrid-api-key \
            --data-file=- \
            --replication-policy="user-managed" \
            --locations="europe-west3,europe-west1"
        echo "✅ Создан (только EU регионы)"
    else
        echo "⚠️  Пропущено (email уведомления не будут работать)"
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Готово!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Список всех секретов:"
gcloud secrets list --filter="name:zeno-auth" --format="table(name,createTime)"
