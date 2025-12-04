#!/bin/bash
# Run migrations on GCP Cloud SQL

set -e

if ! command -v expect &> /dev/null; then
    echo "❌ expect не установлен. Установите: brew install expect"
    exit 1
fi

PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
INSTANCE_ID="${INSTANCE_ID:-zeno-auth-db-dev}"
DB_NAME="${DB_NAME:-zeno_auth}"
DB_USER="${DB_USER:-zeno_auth}"

echo "🔄 Running migrations on GCP Cloud SQL..."
echo "Project: $PROJECT_ID"
echo "Instance: $INSTANCE_ID"
echo "Database: $DB_NAME"
echo ""

MIGRATION_FILE="migrations/001_init_schema.up.sql"

if [ ! -f "$MIGRATION_FILE" ]; then
    echo "❌ Migration file not found: $MIGRATION_FILE"
    exit 1
fi

echo "📦 Applying migration: $MIGRATION_FILE"
echo ""

# Run migration via expect
expect << 'EXPECT_EOF'
set timeout 60
spawn gcloud beta sql connect zeno-auth-db-dev --user=zeno_auth --database=zeno_auth --project=zeno-cy-dev-001
expect "Password:"
send "zte@knp6VXK3xrf3evy\r"
expect "zeno_auth=>"
send "\\i migrations/001_init_schema.up.sql\r"
expect "zeno_auth=>"
send "\\dt\r"
expect "zeno_auth=>"
send "\\q\r"
expect eof
EXPECT_EOF

echo ""
echo "✅ Migration completed successfully!"
