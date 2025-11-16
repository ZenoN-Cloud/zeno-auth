#!/bin/bash
# Скрипт для очистки базы данных через API

API_URL="${1:-https://zeno-auth-dev-174992989924.europe-west3.run.app}"

echo "🗑️  Очистка базы данных через $API_URL/debug/cleanup"
echo "⚠️  Это удалит ВСЕ данные!"
read -p "Продолжить? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Отменено"
    exit 0
fi

curl -X POST "$API_URL/debug/cleanup" \
    -H "Content-Type: application/json" \
    -H "X-Admin-Secret: ${ADMIN_SECRET:-dev-secret}" \
    | jq .

echo ""
echo "✅ Готово"
