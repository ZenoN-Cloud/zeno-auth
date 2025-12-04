#!/bin/bash
# Скрипт для очистки базы данных в GCP

set -e

# Check if expect is installed
if ! command -v expect &> /dev/null; then
    echo "❌ expect не установлен. Установите: brew install expect"
    exit 1
fi

PROJECT_ID="${PROJECT_ID:-zeno-cy-dev-001}"
INSTANCE_ID="${INSTANCE_ID:-zeno-auth-db-dev}"
DB_NAME="${DB_NAME:-zeno_auth}"
DB_USER="${DB_USER:-zeno_auth}"
DB_PASSWORD="${DB_PASSWORD:-zte@knp6VXK3xrf3evy}"

echo "🗑️  Очистка базы данных в GCP Cloud SQL"
echo "Project: $PROJECT_ID"
echo "Instance: $INSTANCE_ID"
echo "Database: $DB_NAME"
echo ""
echo "⚠️  Это удалит ВСЕ данные из таблиц!"
echo ""

read -p "Продолжить? (yes/no): " -r confirm
confirm=$(echo "$confirm" | tr -d '[:space:]')

if [ "$confirm" != "yes" ]; then
    echo "Отменено (введено: '$confirm')"
    exit 0
fi

echo ""
echo "🧹 Очистка таблиц..."

# Use expect to automate password input
expect << 'EXPECT_EOF'
set timeout 30
spawn gcloud beta sql connect zeno-auth-db-dev --user=zeno_auth --database=zeno_auth --project=zeno-cy-dev-001
expect "Password:"
send "zte@knp6VXK3xrf3evy\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE audit_logs CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE user_consents CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE password_reset_tokens CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE email_verifications CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE refresh_tokens CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE org_memberships CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE organizations CASCADE;\r"
expect "zeno_auth=>"
send "TRUNCATE TABLE users CASCADE;\r"
expect "zeno_auth=>"
send "\\q\r"
expect eof
EXPECT_EOF

echo ""
echo "✅ Все таблицы очищены"
echo "✅ Очистка базы данных завершена успешно!"
