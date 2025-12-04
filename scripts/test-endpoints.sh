#!/bin/bash

BASE_URL="${1:-http://localhost:8080}"

# Check required tools
if ! command -v curl >/dev/null 2>&1; then
    echo "❌ curl is required but not installed"
    exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
    echo "❌ jq is required but not installed"
    exit 1
fi

# Validate BASE_URL
if [ -z "$BASE_URL" ]; then
    echo "❌ BASE_URL cannot be empty"
    exit 1
fi

echo "🧪 Тестирование эндпоинтов: $BASE_URL"
echo ""

echo "1️⃣ Health check..."
if curl -s -X GET "$BASE_URL/health" 2>/dev/null | jq '.' 2>/dev/null; then
    echo "✅ Success"
else
    echo "❌ Failed"
fi
echo ""

echo "2️⃣ JWKS endpoint..."
curl -s -X GET "$BASE_URL/jwks" | jq '.' || echo "❌ Failed"
echo ""

echo "3️⃣ Debug endpoint..."
curl -s -X GET "$BASE_URL/debug" | jq '.' || echo "❌ Failed"
echo ""

echo "4️⃣ Register (должен вернуть 400 без данных)..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{}' | jq '.' || echo "❌ Failed"
echo ""

echo "5️⃣ Register с валидными данными..."
EMAIL="test-$(date +%s)@example.com"
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"testpass123\",\"full_name\":\"Test User\"}" | jq '.' || echo "❌ Failed"
echo ""

echo "6️⃣ Login с теми же данными..."
curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"testpass123\"}" | jq '.' || echo "❌ Failed"
echo ""

echo "✅ Тестирование завершено"
