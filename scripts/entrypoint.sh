#!/bin/sh
set -e

echo "Starting zeno-auth..."

# Run migrations if DATABASE_URL is set
if [ -n "$DATABASE_URL" ]; then
    echo "Running database migrations..."
    migrate -path ./migrations -database "$DATABASE_URL" up || {
        echo "Migration failed, but continuing..."
    }
else
    echo "DATABASE_URL not set, skipping migrations"
fi

# Start the application
exec ./zeno-auth
