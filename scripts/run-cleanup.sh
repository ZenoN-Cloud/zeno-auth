#!/bin/bash
set -e

# Run cleanup job
# Usage: ./scripts/run-cleanup.sh [retention-days]
# Default retention: 730 days (2 years)

RETENTION_DAYS=${1:-730}

echo "Running cleanup job with retention: $RETENTION_DAYS days"

cd "$(dirname "$0")/.."

# Build cleanup binary
go build -o ./bin/cleanup ./cmd/cleanup/main.go

# Run cleanup
./bin/cleanup -retention-days=$RETENTION_DAYS

echo "Cleanup completed"
