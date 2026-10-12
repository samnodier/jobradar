#!/usr/bin/env bash
# Pull-based continuous deployment for a single always-on host.
# Run from a checkout of `main` that contains the production `.env`.
# Safe to run repeatedly: it deploys only when main or the published image changed.
set -euo pipefail

cd "$(dirname "$0")/.."
exec 9>/tmp/jobradar-update.lock
flock -n 9 || { echo "another update is running"; exit 0; }

compose() { docker compose -p jobradar -f compose.production.yml "$@"; }
# Host port the app is published on (JOBRADAR_PORT in .env, default 8080).
port=$(sed -n 's/^JOBRADAR_PORT=//p' .env 2>/dev/null | tail -1)
port="${port:-8080}"
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

mkdir -p deploy-state
write_state() {
  printf '{"status":"%s","target_revision":"%s","checked_at":"%s"}\n' "$1" "${release:-unknown}" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > deploy-state/update.json.tmp
  mv deploy-state/update.json.tmp deploy-state/update.json
}
# Persist a sanitized outcome for System health, never command output or secrets.
trap 'write_state failed' ERR

# Pin BOTH artifacts to the selected release. If CI has not published either image,
# leave the running app untouched; the next timer tick retries this same release.
release=$(git rev-parse HEAD)
pinned=$(sed -n 's/^JOBRADAR_TAG=//p' .env | tail -1)
if [ -n "$pinned" ] && [ "$pinned" != main ]; then
  release="$pinned"
fi
export JOBRADAR_TAG="$release"
if ! compose pull --quiet app || ! compose --profile tools pull --quiet migrate; then
  write_state waiting_for_images
  log "release artifacts unavailable; keeping the current application running"
  exit 1
fi
app_image=$(compose config --images app | grep -v -e '^postgres' -e '^redis')
wanted=$(docker image inspect --format '{{.Id}}' "$app_image")
running=$(docker inspect --format '{{.Image}}' "$(compose ps -q app 2>/dev/null)" 2>/dev/null || true)
if [ "$wanted" != "$running" ]; then
  reason="${reason:+$reason; }new application image"
fi

if [ -z "$reason" ]; then
  write_state healthy
  exit 0
fi
write_state deploying
log "deploying: $reason"

# 3. Back up, migrate, restart, verify.
compose up -d db redis
if compose ps -q db >/dev/null && [ -n "$running" ]; then
  ./deploy/backup.sh
fi
# Stop the old app before migrating so old code never runs against a newer schema, and
# never processes work a migration queues (for example re-scoring) with outdated logic.
# Costs a few seconds of downtime per deploy, acceptable for a single personal instance.
compose stop app
compose --profile tools run --rm migrate up
compose up -d app

for _ in $(seq 1 30); do
  health=$(curl --max-time 5 -fsS "http://127.0.0.1:${port}/api/health" || true)
  if printf '%s' "$health" | grep -Eq '"revision"[[:space:]]*:[[:space:]]*"'"$release"'"'; then
    write_state healthy
    log "healthy on running release $release"
    docker image prune -f >/dev/null
    exit 0
  fi
  sleep 2
done
write_state failed
log "app did not become healthy; recent logs:"
compose logs --tail 50 app
exit 1
