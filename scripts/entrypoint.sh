#!/bin/sh
set -e

echo "Starting zeno-auth..."
echo "ENV: $ENV"
echo "PORT: $PORT"
echo "APP_NAME: $APP_NAME"
echo "DATABASE_URL set: $([ -n "$DATABASE_URL" ] && echo 'yes' || echo 'no')"
echo "JWT_PRIVATE_KEY set: $([ -n "$JWT_PRIVATE_KEY" ] && echo 'yes' || echo 'no')"

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
echo "Executing: ./zeno-auth"
exec ./zeno-auth
