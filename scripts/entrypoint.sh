#!/bin/sh
set -e

echo "Starting zeno-auth..."

# Run migrations if DATABASE_URL is set
if [ -n "$DATABASE_URL" ]; then
    echo "Running database migrations..."
    if ! migrate -path ./migrations -database "$DATABASE_URL" up; then
        echo "ERROR: Migration failed! Exiting..."
        exit 1
    fi
    echo "Migrations completed successfully"
else
    echo "WARNING: DATABASE_URL not set, skipping migrations"
fi

# Start the application
echo "Starting application..."
exec ./zeno-auth
