#!/bin/sh
set -e

host="${DB_HOST:-mysql}"
port="${DB_PORT:-3306}"

echo "Waiting for MySQL at ${host}:${port}..."

i=0
while [ "$i" -lt 60 ]; do
  if nc -z "$host" "$port" 2>/dev/null; then
    echo "MySQL is available"
    echo "Running database migrations..."
    ./migrate
    echo "Starting server..."
    exec ./server
  fi
  i=$((i + 1))
  sleep 2
done

echo "MySQL not available after 120 seconds"
exit 1
