#!/bin/sh
set -e

# Apply database migrations before starting the server. Retry a few times so a
# not-yet-ready database does not crash the container.
tries=0
until /app/goose -dir /app/db/migrations postgres "$DATABASE_URL" up; do
  tries=$((tries + 1))
  if [ "$tries" -ge 10 ]; then
    echo "migrations failed after $tries attempts" >&2
    exit 1
  fi
  echo "migrations failed, retrying in 3s..." >&2
  sleep 3
done

exec /app/api
