#!/bin/bash

# Development Data Seeding Script for Zeno Auth

set -e

API_BASE="http://localhost:8080"
ADMIN_EMAIL="admin@zeno.dev"
ADMIN_PASSWORD="AdminPass123!"

# Validate required variables
if [ -z "$API_BASE" ] || [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo "❌ Required variables not set"
    exit 1
fi

# Check required tools
if ! command -v curl >/dev/null 2>&1; then
    echo "❌ curl is required but not installed"
    exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
    echo "❌ jq is required but not installed"
    exit 1
fi

echo "🌱 Seeding development data for Zeno Auth..."

# Check if API is running
if ! curl -s "$API_BASE/health" > /dev/null 2>&1; then
    echo "❌ API is not running at $API_BASE"
    echo "Please start the service with: make local-up"
    exit 1
fi

echo "✅ API is running"

# Function to register a user
register_user() {
    local email=$1
    local password=$2
    local full_name=$3
    
    echo "📝 Registering user: $email"
    
    response=$(curl -s -X POST "$API_BASE/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\",
            \"full_name\": \"$full_name\"
        }")
    
    if echo "$response" | grep -q '"status":"ok"'; then
        echo "✅ User registered: $email"
        return 0
    else
        echo "⚠️  User registration failed or user already exists: $email"
        echo "Response: $response"
        return 1
    fi
}

# Function to login and get tokens
login_user() {
    local email=$1
    local password=$2
    
    echo "🔐 Logging in user: $email"
    
    response=$(curl -s -X POST "$API_BASE/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\"
        }")
    
    if echo "$response" | grep -q '"status":"ok"'; then
        access_token=$(echo "$response" | jq -r '.data.access_token')
        echo "✅ User logged in: $email"
        echo "$access_token"
        return 0
    else
        echo "❌ Login failed: $email"
        echo "Response: $response"
        return 1
    fi
}

# Function to grant consent
grant_consent() {
    local token=$1
    local consent_type=$2
    
    echo "📋 Granting consent: $consent_type"
    
    curl -s -X POST "$API_BASE/v1/me/consents" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"consent_type\": \"$consent_type\",
            \"version\": \"1.0\"
        }" > /dev/null
    
    echo "✅ Consent granted: $consent_type"
}

echo ""
echo "👥 Creating test users..."

# Create admin user
register_user "$ADMIN_EMAIL" "$ADMIN_PASSWORD" "Admin User"
admin_token=$(login_user "$ADMIN_EMAIL" "$ADMIN_PASSWORD")

# Create regular users
register_user "john.doe@example.com" "UserPass123!" "John Doe"
john_token=$(login_user "john.doe@example.com" "UserPass123!")

register_user "jane.smith@example.com" "UserPass123!" "Jane Smith"
jane_token=$(login_user "jane.smith@example.com" "UserPass123!")

register_user "bob.wilson@example.com" "UserPass123!" "Bob Wilson"
bob_token=$(login_user "bob.wilson@example.com" "UserPass123!")

register_user "alice.brown@example.com" "UserPass123!" "Alice Brown"
alice_token=$(login_user "alice.brown@example.com" "UserPass123!")

echo ""
echo "📋 Granting consents for users..."

# Grant consents for users
if [ -n "$john_token" ]; then
    grant_consent "$john_token" "data_processing"
    grant_consent "$john_token" "marketing"
fi

if [ -n "$jane_token" ]; then
    grant_consent "$jane_token" "data_processing"
    grant_consent "$jane_token" "analytics"
fi

if [ -n "$bob_token" ]; then
    grant_consent "$bob_token" "data_processing"
fi

if [ -n "$alice_token" ]; then
    grant_consent "$alice_token" "data_processing"
    grant_consent "$alice_token" "marketing"
    grant_consent "$alice_token" "analytics"
fi

echo ""
echo "🎯 Creating test sessions..."

# Create additional sessions for testing
for user_email in "john.doe@example.com" "jane.smith@example.com"; do
    echo "🔄 Creating additional session for: $user_email"
    login_user "$user_email" "UserPass123!" > /dev/null
done

echo ""
echo "✅ Development data seeding completed!"
echo ""
echo "📊 Summary:"
echo "  • 5 test users created"
echo "  • Multiple consents granted"
echo "  • Test sessions created"
echo ""
echo "🔑 Test Credentials:"
echo "  Admin: $ADMIN_EMAIL / $ADMIN_PASSWORD"
echo "  User1: john.doe@example.com / UserPass123!"
echo "  User2: jane.smith@example.com / UserPass123!"
echo "  User3: bob.wilson@example.com / UserPass123!"
echo "  User4: alice.brown@example.com / UserPass123!"
echo ""
echo "🌐 Access Points:"
echo "  • API: $API_BASE"
echo "  • Health: $API_BASE/health"
echo "  • Metrics: $API_BASE/metrics"
echo "  • JWKS: $API_BASE/.well-known/jwks.json"
echo ""
echo "🧪 Test Commands:"
echo "  make health     # Check API health"
echo "  make metrics    # View metrics"
echo "  make local-test # Run integration tests"