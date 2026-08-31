#!/usr/bin/env sh
set -eu
: "${MYSQL_HOST:?}"
: "${MYSQL_DATABASE:?}"
: "${MYSQL_USER:?}"
: "${MYSQL_PASSWORD:?}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
mkdir -p "$BACKUP_DIR"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
FILE="$BACKUP_DIR/${MYSQL_DATABASE}_${STAMP}.sql.gz"
mysqldump --single-transaction --quick --routines --events -h "$MYSQL_HOST" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" | gzip > "$FILE"
find "$BACKUP_DIR" -type f -name "*.sql.gz" -mtime +"$RETENTION_DAYS" -delete
echo "Backup created: $FILE"
