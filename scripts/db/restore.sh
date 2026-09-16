#!/usr/bin/env bash
# Restores a gzip'd plain-SQL dump (as produced by backup.sh) into the ERP
# database, dropping and recreating it first. Destructive by design --
# requires typing the database name to confirm unless -y/--yes is passed.
#
# Usage:
#   ./restore.sh                          # interactively pick a backup from BACKUP_DIR
#   ./restore.sh backups/clothing_erp_20260101_020000.sql.gz
#   ./restore.sh backups/clothing_erp_20260101_020000.sql.gz --yes   # non-interactive
#
# Environment (all optional, defaults shown):
#   POSTGRES_HOST=localhost
#   POSTGRES_PORT=5432
#   POSTGRES_USER=postgres
#   POSTGRES_PASSWORD=postgres
#   POSTGRES_DB=clothing_erp
#   BACKUP_DIR=<repo>/scripts/db/backups
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-clothing_erp}"
export PGPASSWORD="${POSTGRES_PASSWORD:-postgres}"

BACKUP_DIR="${BACKUP_DIR:-$SCRIPT_DIR/backups}"

usage() {
  sed -n '2,15p' "${BASH_SOURCE[0]}" | sed 's/^# \?//'
}

FORCE=0
FILE=""
for arg in "$@"; do
  case "$arg" in
    -y|--yes) FORCE=1 ;;
    -h|--help) usage; exit 0 ;;
    *) FILE="$arg" ;;
  esac
done

if [ -z "$FILE" ]; then
  echo "No backup file given. Available backups in $BACKUP_DIR:"
  shopt -s nullglob
  options=("$BACKUP_DIR"/*.sql.gz)
  shopt -u nullglob
  if [ "${#options[@]}" -eq 0 ]; then
    echo "No backups found in $BACKUP_DIR" >&2
    exit 1
  fi
  select opt in "${options[@]}"; do
    if [ -n "$opt" ]; then FILE="$opt"; break; fi
  done
fi

if [ ! -f "$FILE" ]; then
  echo "Backup file not found: $FILE" >&2
  exit 1
fi

echo "About to restore into database '$POSTGRES_DB' on $POSTGRES_HOST:$POSTGRES_PORT"
echo "Source: $FILE"
echo
echo "!! THIS WILL PERMANENTLY DROP AND REPLACE ALL DATA IN THAT DATABASE !!"

if [ "$FORCE" -ne 1 ]; then
  read -r -p "Type the database name ($POSTGRES_DB) to confirm: " confirm
  if [ "$confirm" != "$POSTGRES_DB" ]; then
    echo "Confirmation did not match. Aborting." >&2
    exit 1
  fi
fi

echo "[restore] terminating other connections to '$POSTGRES_DB'..."
psql --host="$POSTGRES_HOST" --port="$POSTGRES_PORT" --username="$POSTGRES_USER" --dbname=postgres -v ON_ERROR_STOP=1 \
  -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$POSTGRES_DB' AND pid <> pg_backend_pid();"

echo "[restore] dropping and recreating '$POSTGRES_DB'..."
psql --host="$POSTGRES_HOST" --port="$POSTGRES_PORT" --username="$POSTGRES_USER" --dbname=postgres -v ON_ERROR_STOP=1 \
  -c "DROP DATABASE IF EXISTS \"$POSTGRES_DB\";" \
  -c "CREATE DATABASE \"$POSTGRES_DB\";"

echo "[restore] loading $FILE..."
gunzip -c "$FILE" | psql --host="$POSTGRES_HOST" --port="$POSTGRES_PORT" --username="$POSTGRES_USER" --dbname="$POSTGRES_DB" -v ON_ERROR_STOP=1

echo "[restore] complete."
