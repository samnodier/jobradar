#!/usr/bin/env bash
# Run on your development machine. Keep secrets/database dumps outside the repository.
set -euo pipefail
if [ "$#" -ne 2 ]; then
  echo 'Usage: deploy/copy-backup.sh SSH_HOST LOCAL_BACKUP_DIRECTORY' >&2
  exit 2
fi
host="$1"
dir="$2"
[[ "$host" =~ ^[a-zA-Z0-9_.@-]+$ && "$host" != -* ]] || { echo 'Invalid SSH host' >&2; exit 2; }
mkdir -p "$dir"
chmod 700 "$dir"
# Have the server choose its newest completed dump; no remote path is evaluated locally.
ssh -o BatchMode=yes "$host" 'bash -c '\''set -euo pipefail; file=$(find "$HOME/jobradar-backups" -maxdepth 1 -name "jobradar-*.dump" -type f | sort | tail -1); test -n "$file"; cat "$file"'\''' > "$dir/latest.dump.partial"
test -s "$dir/latest.dump.partial"
chmod 600 "$dir/latest.dump.partial"
mv "$dir/latest.dump.partial" "$dir/jobradar-$(date -u +%Y%m%d-%H%M%S).dump"
echo 'Backup copied. Keep your stable encryption key in a separate private recovery record.'
