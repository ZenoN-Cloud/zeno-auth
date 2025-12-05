#!/bin/bash
# Проверка версий миграций в GCP базе

set -e

if ! command -v expect &> /dev/null; then
    echo "❌ expect не установлен"
    exit 1
fi

echo "🔍 Проверка версий миграций в GCP Cloud SQL..."

expect << 'EXPECT_EOF'
set timeout 30
spawn gcloud beta sql connect zeno-auth-db-dev --user=zeno_auth --database=zeno_auth --project=zeno-cy-dev-001
expect "Password:"
send "zte@knp6VXK3xrf3evy\r"
expect "zeno_auth=>"
send "SELECT * FROM schema_migrations ORDER BY version;\r"
expect "zeno_auth=>"
send "\\q\r"
expect eof
EXPECT_EOF
