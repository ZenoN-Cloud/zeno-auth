#!/bin/bash
# Сброс версий миграций в GCP базе

set -e

if ! command -v expect &> /dev/null; then
    echo "❌ expect не установлен. Установите: brew install expect"
    exit 1
fi

echo "🔄 Сброс версий миграций в GCP Cloud SQL..."
echo ""
echo "⚠️  Это удалит записи о старых миграциях из goose_db_version!"
echo ""

read -p "Продолжить? (yes/no): " -r confirm
confirm=$(echo "$confirm" | tr -d '[:space:]')

if [ "$confirm" != "yes" ]; then
    echo "Отменено"
    exit 0
fi

echo ""
echo "🧹 Очистка версий миграций..."

# Use expect to automate password input
expect << 'EXPECT_EOF'
set timeout 30
spawn gcloud beta sql connect zeno-auth-db-dev --user=zeno_auth --database=zeno_auth --project=zeno-cy-dev-001
expect "Password:"
send "zte@knp6VXK3xrf3evy\r"
expect "zeno_auth=>"
send "DROP TABLE IF EXISTS schema_migrations CASCADE;\r"
expect "zeno_auth=>"
send "\\q\r"
expect eof
EXPECT_EOF

echo ""
echo "✅ Таблица goose_db_version удалена!"
echo "Теперь можно задеплоить заново"
