#!/usr/bin/env bash
# Dumps the ERP database to a timestamped, gzip-compressed plain-SQL file
# and prunes backups older than RETENTION_DAYS. Plain SQL (not pg_dump's
# custom -Fc format) is used deliberately: it's restorable with plain psql
# (see restore.sh), diffable, and doesn't tie you to a matching pg_restore
# version.
#
# Usage:
#   ./backup.sh                    # uses defaults / environment below
#   POSTGRES_HOST=db.internal ./backup.sh
#
# Environment (all optional, defaults shown):
#   POSTGRES_HOST=localhost
#   POSTGRES_PORT=5432
#   POSTGRES_USER=postgres
#   POSTGRES_PASSWORD=postgres
#   POSTGRES_DB=clothing_erp
#   BACKUP_DIR=<repo>/scripts/db/backups
#   RETENTION_DAYS=30
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-clothing_erp}"
export PGPASSWORD="${POSTGRES_PASSWORD:-postgres}"

BACKUP_DIR="${BACKUP_DIR:-$SCRIPT_DIR/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

mkdir -p "$BACKUP_DIR"

timestamp="$(date +%Y%m%d_%H%M%S)"
out_file="$BACKUP_DIR/clothing_erp_${timestamp}.sql.gz"
tmp_file="${out_file}.tmp"

echo "[backup] dumping ${POSTGRES_DB}@${POSTGRES_HOST}:${POSTGRES_PORT} -> ${out_file}"

pg_dump \
  --host="$POSTGRES_HOST" \
  --port="$POSTGRES_PORT" \
  --username="$POSTGRES_USER" \
  --dbname="$POSTGRES_DB" \
  --format=plain \
  --no-owner \
  --no-privileges \
  | gzip -9 > "$tmp_file"

# Only replace the final filename once the dump has fully succeeded, so a
# failed/interrupted run never leaves a truncated file with a "real" name
# that a later restore could pick up by mistake.
mv "$tmp_file" "$out_file"

echo "[backup] done: $(du -h "$out_file" | cut -f1)"

echo "[backup] pruning backups older than ${RETENTION_DAYS} days in ${BACKUP_DIR}"
find "$BACKUP_DIR" -name 'clothing_erp_*.sql.gz' -type f -mtime "+${RETENTION_DAYS}" -print -delete

echo "[backup] complete."
