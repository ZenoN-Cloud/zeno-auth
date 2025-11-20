#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_URL="http://localhost:8080"

echo "🔒 Testing Zeno Auth Security Features"
echo "========================================"
echo ""

# Test 1: Health Check
echo "1️⃣  Testing Health Check..."
response=$(curl -s -o /dev/null -w "%{http_code}" $API_URL/health)
if [ "$response" -eq 200 ]; then
    echo -e "${GREEN}✅ Health check passed${NC}"
else
    echo -e "${RED}❌ Health check failed (HTTP $response)${NC}"
fi
echo ""

# Test 2: Password Validation - Weak Password
echo "2️⃣  Testing Password Validation (weak password)..."
response=$(curl -s -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test1@example.com","password":"weak","full_name":"Test User"}')
if echo "$response" | grep -q "password"; then
    echo -e "${GREEN}✅ Weak password rejected${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}❌ Weak password accepted (should be rejected)${NC}"
    echo "   Response: $response"
fi
echo ""

# Test 3: Password Validation - No Uppercase
echo "3️⃣  Testing Password Validation (no uppercase)..."
response=$(curl -s -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@example.com","password":"lowercase123","full_name":"Test User"}')
if echo "$response" | grep -q "uppercase"; then
    echo -e "${GREEN}✅ Password without uppercase rejected${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}❌ Password without uppercase accepted${NC}"
    echo "   Response: $response"
fi
echo ""

# Test 4: Password Validation - Common Password
echo "4️⃣  Testing Password Validation (common password)..."
response=$(curl -s -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test3@example.com","password":"password123","full_name":"Test User"}')
if echo "$response" | grep -q "common"; then
    echo -e "${GREEN}✅ Common password rejected${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}❌ Common password accepted${NC}"
    echo "   Response: $response"
fi
echo ""

# Test 5: Input Validation - Invalid Email
echo "5️⃣  Testing Input Validation (invalid email)..."
response=$(curl -s -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"not-an-email","password":"SecurePass123","full_name":"Test User"}')
if echo "$response" | grep -q "email"; then
    echo -e "${GREEN}✅ Invalid email rejected${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}❌ Invalid email accepted${NC}"
    echo "   Response: $response"
fi
echo ""

# Test 6: Valid Registration
echo "6️⃣  Testing Valid Registration..."
response=$(curl -s -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"valid@example.com","password":"SecurePass123","full_name":"Valid User"}')
if echo "$response" | grep -q "id"; then
    echo -e "${GREEN}✅ Valid registration successful${NC}"
    echo "   Response: $response"
else
    echo -e "${RED}❌ Valid registration failed${NC}"
    echo "   Response: $response"
fi
echo ""

# Test 7: Rate Limiting on Login
echo "7️⃣  Testing Rate Limiting (6 login attempts)..."
echo "   Sending 6 requests rapidly..."
rate_limited=false
for i in {1..6}; do
    response=$(curl -s -w "\n%{http_code}" -X POST $API_URL/auth/login \
      -H "Content-Type: application/json" \
      -d '{"email":"test@example.com","password":"wrong"}')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" -eq 429 ]; then
        rate_limited=true
        echo -e "   Request $i: ${YELLOW}HTTP 429 - Rate Limited${NC}"
        break
    else
        echo "   Request $i: HTTP $http_code"
    fi
done

if [ "$rate_limited" = true ]; then
    echo -e "${GREEN}✅ Rate limiting working${NC}"
else
    echo -e "${YELLOW}⚠️  Rate limit not triggered (may need more requests)${NC}"
fi
echo ""

# Test 8: Security Headers
echo "8️⃣  Testing Security Headers..."
headers=$(curl -s -I $API_URL/health)

check_header() {
    header_name=$1
    if echo "$headers" | grep -qi "$header_name"; then
        echo -e "   ${GREEN}✅ $header_name present${NC}"
    else
        echo -e "   ${RED}❌ $header_name missing${NC}"
    fi
}

check_header "X-Frame-Options"
check_header "X-Content-Type-Options"
check_header "Strict-Transport-Security"
check_header "Content-Security-Policy"
echo ""

# Test 9: CORS Headers
echo "9️⃣  Testing CORS Headers..."
cors_response=$(curl -s -I -H "Origin: http://localhost:5173" $API_URL/health)
if echo "$cors_response" | grep -qi "Access-Control-Allow-Origin"; then
    echo -e "${GREEN}✅ CORS headers present${NC}"
    echo "$cors_response" | grep -i "Access-Control"
else
    echo -e "${RED}❌ CORS headers missing${NC}"
fi
echo ""

echo "========================================"
echo "🎉 Security Testing Complete!"
echo ""
