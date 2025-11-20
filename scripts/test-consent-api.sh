#!/bin/bash

set -e

API_URL="${API_URL:-http://localhost:8080}"
EMAIL="consent-test@example.com"
PASSWORD="TestPass123!"

echo "🧪 Testing Consent Management API"
echo "=================================="
echo ""

# 1. Register user
echo "1️⃣  Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"full_name\": \"Consent Test User\"
  }")

echo "Response: $REGISTER_RESPONSE"
echo ""

# 2. Login
echo "2️⃣  Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Failed to get access token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Got access token: ${ACCESS_TOKEN:0:20}..."
echo ""

# 3. Grant Terms consent
echo "3️⃣  Granting Terms consent..."
GRANT_TERMS=$(curl -s -X POST "$API_URL/me/consents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "consent_type": "terms",
    "version": "1.0"
  }')

echo "Response: $GRANT_TERMS"
echo ""

# 4. Grant Privacy consent
echo "4️⃣  Granting Privacy consent..."
GRANT_PRIVACY=$(curl -s -X POST "$API_URL/me/consents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "consent_type": "privacy",
    "version": "1.0"
  }')

echo "Response: $GRANT_PRIVACY"
echo ""

# 5. Grant Marketing consent
echo "5️⃣  Granting Marketing consent..."
GRANT_MARKETING=$(curl -s -X POST "$API_URL/me/consents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "consent_type": "marketing",
    "version": "1.0"
  }')

echo "Response: $GRANT_MARKETING"
echo ""

# 6. Get all consents
echo "6️⃣  Getting all consents..."
GET_CONSENTS=$(curl -s -X GET "$API_URL/me/consents" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response: $GET_CONSENTS"
echo ""

# 7. Revoke Marketing consent
echo "7️⃣  Revoking Marketing consent..."
REVOKE_MARKETING=$(curl -s -X DELETE "$API_URL/me/consents/marketing" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response: $REVOKE_MARKETING"
echo ""

# 8. Get consents again (should not include marketing)
echo "8️⃣  Getting consents after revocation..."
GET_CONSENTS_AFTER=$(curl -s -X GET "$API_URL/me/consents" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response: $GET_CONSENTS_AFTER"
echo ""

# 9. Update consent version (grant new version)
echo "9️⃣  Updating Terms consent to version 2.0..."
UPDATE_TERMS=$(curl -s -X POST "$API_URL/me/consents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "consent_type": "terms",
    "version": "2.0"
  }')

echo "Response: $UPDATE_TERMS"
echo ""

# 10. Final check
echo "🔟 Final consent check..."
FINAL_CONSENTS=$(curl -s -X GET "$API_URL/me/consents" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response: $FINAL_CONSENTS"
echo ""

echo "✅ Consent Management API tests completed!"
