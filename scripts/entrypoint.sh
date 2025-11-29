#!/bin/sh
set -e

echo "========================================"
echo "  Starting zeno-auth"
echo "========================================"
echo "ENV:               ${ENV}"
echo "PORT:              ${PORT}"
echo "APP_NAME:          ${APP_NAME}"
echo "DATABASE_URL:      $([ -n "${DATABASE_URL}" ] && echo 'set' || echo 'NOT set')"
echo "JWT_PRIVATE_KEY:   $([ -n "${JWT_PRIVATE_KEY}" ] && echo 'set' || echo 'NOT set')"
echo "----------------------------------------"

# ---- WAIT FOR DATABASE -------------------------------------------------------
if [ -n "${DATABASE_URL}" ]; then
  echo "Waiting for database to become available..."
  RETRIES=30
  SLEEP=3

  for i in $(seq 1 $RETRIES); do
    if migrate -path ./migrations -database "${DATABASE_URL}" version 2>&1 | grep -qE "(no migration|^[0-9])"; then
      echo "Database is reachable!"
      break
    fi

    echo "Attempt ${i}/${RETRIES}... database not ready yet"
    sleep $SLEEP
  done

  # Check after final attempt
  if ! migrate -path ./migrations -database "${DATABASE_URL}" version 2>&1 | grep -qE "(no migration|^[0-9])"; then
    echo "ERROR: Database is still unreachable after ${RETRIES} attempts"
    exit 1
  fi
else
  echo "WARNING: DATABASE_URL not set — skipping DB wait"
fi

# ---- RUN MIGRATIONS ----------------------------------------------------------
if [ -n "${DATABASE_URL}" ]; then
  echo "Running migrations..."
  if migrate -path ./migrations -database "${DATABASE_URL}" up; then
    echo "Migrations completed successfully"
  else
    echo "ERROR: Migration failed!"
    exit 1
  fi
else
  echo "Skipping migrations (DATABASE_URL not provided)"
fi

# ---- START APPLICATION --------------------------------------------------------
echo "----------------------------------------"
echo "Starting application: ./zeno-auth"
echo "----------------------------------------"

# exec → replace shell with app (correct for Cloud Run)
if [ ! -f ./zeno-auth ]; then
  echo "ERROR: Binary ./zeno-auth not found!"
  exit 1
fi

if [ ! -x ./zeno-auth ]; then
  echo "ERROR: Binary ./zeno-auth is not executable!"
  exit 1
fi

exec ./zeno-auth
