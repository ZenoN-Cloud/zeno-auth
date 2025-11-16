#!/bin/bash

BASE_URL="${1:-http://localhost:8080}"

echo "🧪 Тестирование эндпоинтов: $BASE_URL"
echo ""

echo "1️⃣ Health check..."
curl -s -X GET "$BASE_URL/health" | jq '.' || echo "❌ Failed"
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
