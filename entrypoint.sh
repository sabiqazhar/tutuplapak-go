#!/bin/bash
set -e

echo "🚀 Applying database migrations..."
./migrate -path ./migrations -database "$DATABASE_URL" up

echo "✅ Migrations completed. Starting application..."
exec ./main