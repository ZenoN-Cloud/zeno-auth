#!/bin/bash
set -e

# Database migration script for Cloud Run
# Usage: ./migrate.sh [up|down] [DATABASE_URL]

COMMAND=${1:-up}
DATABASE_URL=${2:-$DATABASE_URL}

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not provided"
    exit 1
fi

echo "Running migrations: $COMMAND"
echo "Database: $(echo $DATABASE_URL | sed 's/:[^:]*@/@***@/')"

# Install migrate if not present
if ! command -v migrate &> /dev/null; then
    echo "Installing golang-migrate..."
    curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
    sudo mv migrate /usr/local/bin/
fi

# Run migrations
migrate -path migrations -database "$DATABASE_URL" $COMMAND

echo "Migrations completed successfully"