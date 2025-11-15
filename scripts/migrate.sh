#!/bin/bash
set -e

DATABASE_URL="${DATABASE_URL:-}"

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is required"
    exit 1
fi

echo "Running migrations..."

for migration in migrations/*.sql; do
    echo "Applying $(basename $migration)..."
    psql "$DATABASE_URL" -f "$migration"
done

echo "Migrations completed successfully!"
