#!/usr/bin/env bash
# Dump the production database to BACKUP_DIR and keep the newest BACKUP_KEEP dumps.
# Copy these files off the host as well: a backup on the same disk does not survive a dead disk.
set -euo pipefail

cd "$(dirname "$0")/.."
dir="${BACKUP_DIR:-$HOME/jobradar-backups}"
keep="${BACKUP_KEEP:-14}"
mkdir -p "$dir"
chmod 700 "$dir"

file="$dir/jobradar-$(date +%Y%m%d-%H%M%S).dump"
docker compose -p jobradar -f compose.production.yml exec -T db \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc --no-owner --no-acl' > "$file.partial"
if [ ! -s "$file.partial" ]; then
  rm -f "$file.partial"
  echo "[backup] dump was empty" >&2
  exit 1
fi
mv "$file.partial" "$file"
chmod 600 "$file"
echo "[backup] wrote $file ($(du -h "$file" | cut -f1))"

ls -1t "$dir"/jobradar-*.dump | tail -n +"$((keep + 1))" | xargs -r rm -f
