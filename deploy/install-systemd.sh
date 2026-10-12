#!/usr/bin/env bash
# Install user-level systemd timers for automatic updates (every 5 minutes) and nightly backups.
# Requires: the current user can run docker, and lingering is enabled so timers run without a login:
#   sudo loginctl enable-linger "$USER"
set -euo pipefail

repo="$(cd "$(dirname "$0")/.." && pwd)"
units="$HOME/.config/systemd/user"
mkdir -p "$units"

cat > "$units/jobradar-update.service" <<EOF
[Unit]
Description=Jobradar: deploy new versions from main
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
WorkingDirectory=$repo
ExecStart=$repo/deploy/update.sh
EOF

cat > "$units/jobradar-update.timer" <<EOF
[Unit]
Description=Jobradar: check for new versions every 5 minutes

[Timer]
OnBootSec=2min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
EOF

cat > "$units/jobradar-backup.service" <<EOF
[Unit]
Description=Jobradar: nightly database backup

[Service]
Type=oneshot
WorkingDirectory=$repo
ExecStart=$repo/deploy/backup.sh
EOF

cat > "$units/jobradar-backup.timer" <<EOF
[Unit]
Description=Jobradar: nightly database backup

[Timer]
OnCalendar=*-*-* 03:30
Persistent=true

[Install]
WantedBy=timers.target
EOF

mkdir -p "$repo/deploy-state"
cat > "$units/jobradar-monitor.service" <<EOF
[Unit]
Description=Jobradar: check operations and notify owner
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
WorkingDirectory=$repo
ExecStart=$repo/deploy/monitor.sh
EOF
cat > "$units/jobradar-monitor.timer" <<EOF
[Unit]
Description=Jobradar: operational check every 15 minutes

[Timer]
OnBootSec=10min
OnUnitActiveSec=15min

[Install]
WantedBy=timers.target
EOF
systemctl --user daemon-reload
systemctl --user enable --now jobradar-update.timer jobradar-backup.timer jobradar-monitor.timer
systemctl --user list-timers 'jobradar-*'
loginctl show-user "$USER" -p Linger
