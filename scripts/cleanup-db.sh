#!/bin/bash
# Скрипт для очистки базы данных через API

API_URL="${1:-https://zeno-auth-dev-174992989924.europe-west3.run.app}"

echo "🗑️  Очистка базы данных через $API_URL/debug/cleanup"
echo "⚠️  Это удалит ВСЕ данные!"

# Check API availability
echo "🔍 Проверка доступности API..."
if ! curl -s --max-time 10 "$API_URL/health" > /dev/null; then
    echo "❌ API недоступен: $API_URL"
    echo "Проверьте URL и состояние сервиса"
    exit 1
fi
echo "✅ API доступен"

read -p "Продолжить? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Отменено"
    exit 0
fi

# Execute cleanup with proper error handling
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/debug/cleanup" \
    -H "Content-Type: application/json" \
    -H "X-Admin-Secret: ${ADMIN_SECRET:-dev-secret}")

# Extract response body and status code
http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)

echo "HTTP Status: $http_code"

# Check if request was successful
if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 204 ]; then
    echo "Response:"
    echo "$response_body" | jq . 2>/dev/null || echo "$response_body"
    echo ""
    echo "✅ Очистка базы данных завершена успешно"
else
    echo "❌ Ошибка при очистке базы данных"
    echo "Response:"
    echo "$response_body" | jq . 2>/dev/null || echo "$response_body"
    echo ""
    exit 1
fi
