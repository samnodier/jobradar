#!/usr/bin/env bash
# Separate from the application process: detects app failures while Docker is up.
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p deploy-state
exec 8>/tmp/jobradar-monitor.lock
flock -n 8 || exit 0
# Reuse the RUNNING image, not main's unverified/latest image. Match the host UID
# so the one-off container can write its deduplication markers to the shared folder.
container=$(docker compose -p jobradar -f compose.production.yml ps -aq app)
if [ -z "$container" ]; then echo 'No deployed application image for monitoring' >&2; exit 1; fi
image=$(docker inspect --format '{{.Image}}' "$container")
override=$(mktemp)
trap 'rm -f "$override"' EXIT
printf 'services:\n  app:\n    image: "%s"\n' "$image" > "$override"
docker compose -p jobradar -f compose.production.yml -f "$override" run --rm --no-deps \
  --user "$(id -u):$(id -g)" --entrypoint /app/jobradar \
  -e OPERATIONS_STATE_DIR=/app/deploy-state \
  -e OPS_HEALTH_URL=http://app:8080/api/health \
  app --ops-check
