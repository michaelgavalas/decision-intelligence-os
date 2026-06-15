#!/usr/bin/env bash
# Nightly logical backup of the Postgres database.
# Intended to run from cron on the host, e.g.:
#   0 3 * * * /opt/dios/infra/scripts/backup.sh >> /var/log/dios-backup.log 2>&1
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/var/backups/dios}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
DB_SERVICE="${DB_SERVICE:-postgres}"
COMPOSE_FILE="${COMPOSE_FILE:-infra/docker-compose.yml}"

mkdir -p "$BACKUP_DIR"
OUT="$BACKUP_DIR/dios-$TIMESTAMP.sql.gz"

docker compose -f "$COMPOSE_FILE" exec -T "$DB_SERVICE" \
	pg_dump -U dios -d dios --no-owner --clean --if-exists | gzip >"$OUT"

echo "wrote $OUT"

# Prune old backups.
find "$BACKUP_DIR" -name 'dios-*.sql.gz' -mtime "+$RETENTION_DAYS" -delete
