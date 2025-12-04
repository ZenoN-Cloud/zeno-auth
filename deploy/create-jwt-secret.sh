#!/bin/bash
# Quick script to create JWT private key secret

set -e

echo "🔐 Creating JWT Private Key Secret..."
echo ""

# Generate temporary RSA key
if ! TEMP_KEY=$(mktemp); then
  echo "❌ Failed to create temporary file"
  exit 1
fi

# Ensure cleanup on exit
trap 'rm -f "$TEMP_KEY"' EXIT

if ! openssl genrsa -out "$TEMP_KEY" 2048 2>/dev/null; then
  echo "❌ Failed to generate RSA key"
  exit 1
fi

echo "✅ RSA key generated"

# Create secret
if gcloud secrets create zeno-auth-jwt-private-key \
  --data-file="$TEMP_KEY" \
  --replication-policy="automatic" 2>/dev/null; then
  echo "✅ Secret created: zeno-auth-jwt-private-key"
else
  echo "❌ Failed to create secret (may already exist)"
  echo "To update existing secret, use: gcloud secrets versions add zeno-auth-jwt-private-key --data-file=\"$TEMP_KEY\""
fi

# Cleanup handled by trap

echo ""
echo "Done! You can now deploy with: ./deploy/gcp-deploy.sh"
