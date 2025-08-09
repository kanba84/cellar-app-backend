#!/bin/bash
set -e


echo "Restoring database from dump file..."

pg_restore \
  --no-owner \
  --username="$POSTGRES_USER" \
  --dbname="$POSTGRES_DB" \
  --verbose \
  /docker-entrypoint-initdb.d/cellar_app.dump

echo "Restore completed."
