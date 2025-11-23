#!/bin/bash
# Quick script to create JWT private key secret

set -e

echo "🔐 Creating JWT Private Key Secret..."
echo ""

# Generate temporary RSA key
TEMP_KEY=$(mktemp)
openssl genrsa -out "$TEMP_KEY" 2048 2>/dev/null

echo "✅ RSA key generated"

# Create secret
gcloud secrets create zeno-auth-jwt-private-key \
  --data-file="$TEMP_KEY" \
  --replication-policy="automatic"

echo "✅ Secret created: zeno-auth-jwt-private-key"

# Cleanup
rm -f "$TEMP_KEY"

echo ""
echo "Done! You can now deploy with: ./deploy/gcp-deploy.sh"
