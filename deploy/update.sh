#!/usr/bin/env bash
# Pull-based continuous deployment for a single always-on host.
# Run from a checkout of `main` that contains the production `.env`.
# Safe to run repeatedly: it deploys only when main or the published image changed.
set -euo pipefail

cd "$(dirname "$0")/.."
exec 9>/tmp/jobradar-update.lock
flock -n 9 || { echo "another update is running"; exit 0; }

compose() { docker compose -p jobradar -f compose.production.yml "$@"; }
log() { echo "[update] $*"; }

# A re-exec after pulling main passes along why it is deploying.
reason="${JOBRADAR_UPDATE_REASON:-}"

# 1. Configuration and scripts come from main.
if [ "$(git branch --show-current)" != "main" ]; then
  log "this checkout must be on main; not deploying"
  exit 1
fi
git fetch --quiet origin main
if [ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]; then
  if ! git merge --ff-only --quiet origin/main; then
    log "local checkout has diverged from origin/main; not deploying"
    exit 1
  fi
  # The script itself may have changed; restart with the new version.
  JOBRADAR_UPDATE_REASON="main moved to $(git rev-parse --short HEAD)" exec "$0" "$@"
fi

# 2. Images come from the registry (published by CI after tests pass).
compose pull --quiet app
compose --profile tools pull --quiet migrate
app_image=$(compose config --images app | grep -v -e '^postgres' -e '^redis')
wanted=$(docker image inspect --format '{{.Id}}' "$app_image")
running=$(docker inspect --format '{{.Image}}' "$(compose ps -q app 2>/dev/null)" 2>/dev/null || true)
if [ "$wanted" != "$running" ]; then
  reason="${reason:+$reason; }new application image"
fi

if [ -z "$reason" ]; then
  exit 0
fi
log "deploying: $reason"

# 3. Back up, migrate, restart, verify.
compose up -d db redis
if compose ps -q db >/dev/null && [ -n "$running" ]; then
  ./deploy/backup.sh
fi
compose --profile tools run --rm migrate up
compose up -d app

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8080/api/health >/dev/null; then
    log "healthy on $(git rev-parse --short HEAD)"
    docker image prune -f >/dev/null
    exit 0
  fi
  sleep 2
done
log "app did not become healthy; recent logs:"
compose logs --tail 50 app
exit 1
