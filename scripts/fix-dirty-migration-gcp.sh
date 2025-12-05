#!/bin/bash
# Fix dirty migration state in GCP database

set -e

if ! command -v expect &> /dev/null; then
    echo "❌ expect не установлен"
    exit 1
fi

echo "🔧 Исправление dirty migration state в GCP Cloud SQL..."
echo ""
echo "⚠️  Это сбросит dirty flag в schema_migrations!"
echo ""

read -p "Продолжить? (yes/no): " -r confirm
confirm=$(echo "$confirm" | tr -d '[:space:]')

if [ "$confirm" != "yes" ]; then
    echo "Отменено"
    exit 0
fi

echo ""
echo "🔧 Исправление dirty state..."

expect << 'EXPECT_EOF'
set timeout 30
spawn gcloud beta sql connect zeno-auth-db-dev --user=zeno_auth --database=zeno_auth --project=zeno-cy-dev-001
expect "Password:"
send "zte@knp6VXK3xrf3evy\r"
expect "zeno_auth=>"
send "UPDATE schema_migrations SET dirty = false WHERE version = 1;\r"
expect "zeno_auth=>"
send "SELECT * FROM schema_migrations;\r"
expect "zeno_auth=>"
send "\\q\r"
expect eof
EXPECT_EOF

echo ""
echo "✅ Dirty state исправлен!"
echo "Теперь можно задеплоить заново"
