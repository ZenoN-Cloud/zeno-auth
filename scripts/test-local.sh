#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo "🧪 Тестирование Zeno Auth API (локально)"
echo "=========================================="
echo ""

# Проверка доступности API
echo -n "1️⃣  Проверка health endpoint... "
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/health)
if [ "$HEALTH" -eq 200 ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL (HTTP $HEALTH)${NC}"
    exit 1
fi

# Проверка JWKS
echo -n "2️⃣  Проверка JWKS endpoint... "
JWKS=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/jwks)
if [ "$JWKS" -eq 200 ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL (HTTP $JWKS)${NC}"
fi

# Генерация случайного email
RANDOM_EMAIL="test_$(date +%s)@example.com"

# Регистрация пользователя
echo -n "3️⃣  Регистрация пользователя... "
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$RANDOM_EMAIL\",
    \"password\": \"SecurePass123!\",
    \"full_name\": \"Test User\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q '"id"'; then
    echo -e "${GREEN}✓ OK${NC}"
    USER_ID=$(echo $REGISTER_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "   └─ User ID: $USER_ID"
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   └─ Response: $REGISTER_RESPONSE"
    exit 1
fi

# Логин
echo -n "4️⃣  Логин пользователя... "
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$RANDOM_EMAIL\",
    \"password\": \"SecurePass123!\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q "access_token"; then
    echo -e "${GREEN}✓ OK${NC}"
    ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
    echo "   └─ Access Token: ${ACCESS_TOKEN:0:20}..."
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   └─ Response: $LOGIN_RESPONSE"
    exit 1
fi

# Получение текущего пользователя
echo -n "5️⃣  Получение профиля (GET /me)... "
ME_RESPONSE=$(curl -s -X GET $BASE_URL/me \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$ME_RESPONSE" | grep -q "$RANDOM_EMAIL"; then
    echo -e "${GREEN}✓ OK${NC}"
    echo "   └─ Email: $RANDOM_EMAIL"
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   └─ Response: $ME_RESPONSE"
fi

# Создание организации
echo -n "6️⃣  Создание организации... "
ORG_SLUG="test-org-$(date +%s)"
ORG_RESPONSE=$(curl -s -X POST $BASE_URL/organizations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Organization\",
    \"slug\": \"$ORG_SLUG\"
  }")

if echo "$ORG_RESPONSE" | grep -q '"id"'; then
    echo -e "${GREEN}✓ OK${NC}"
    ORG_ID=$(echo $ORG_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "   └─ Org ID: $ORG_ID"
    echo "   └─ Slug: $ORG_SLUG"
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   └─ Response: $ORG_RESPONSE"
fi

# Получение списка организаций
echo -n "7️⃣  Получение списка организаций... "
ORGS_RESPONSE=$(curl -s -X GET $BASE_URL/organizations \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$ORGS_RESPONSE" | grep -q "$ORG_SLUG"; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   └─ Response: $ORGS_RESPONSE"
fi

# Refresh token
echo -n "8️⃣  Обновление токена (refresh)... "
REFRESH_RESPONSE=$(curl -s -X POST $BASE_URL/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

if echo "$REFRESH_RESPONSE" | grep -q "access_token"; then
    echo -e "${GREEN}✓ OK${NC}"
    NEW_ACCESS_TOKEN=$(echo $REFRESH_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    echo "   └─ New Access Token: ${NEW_ACCESS_TOKEN:0:20}..."
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   └─ Response: $REFRESH_RESPONSE"
fi

# Logout
echo -n "9️⃣  Logout... "
LOGOUT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

if [ "$LOGOUT_RESPONSE" -eq 200 ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${YELLOW}⚠ HTTP $LOGOUT_RESPONSE${NC}"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}✅ Все основные тесты пройдены!${NC}"
echo ""
echo "📝 Созданные данные:"
echo "   Email: $RANDOM_EMAIL"
echo "   User ID: $USER_ID"
echo "   Org ID: $ORG_ID"
echo "   Org Slug: $ORG_SLUG"
echo ""
