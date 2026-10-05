# Session handoff — October 5, 2026

Branch: `dev`. `main` is promoted after `dev` is tested on the new development machine.

- Set up a new machine with `docs/MOVING_MACHINES.md`; data and `.env` travel privately, never through git.
- No scheduled digest had been delivered (development Compose lacked a restart policy; fixed). Only one machine may run the digest scheduler.
- Remaining checks: backend integration suite, `docs/REVIEW_CHECKLIST.md`, a real scheduled digest, Jobicy/MyJobMag live ingestion, PDF print, one week unattended, backup restore.

Next decisions: process supervision on the always-on host (systemd vs `compose.production.yml`), public access, off-host backups, missed-digest alerting.

Working agreement: direct implementation, small focused commits on feature branches off `dev`, LEARNINGS.md maintained, no assistant attribution. Sam runs `make run`/`npm run dev` himself. Keep personal machine, network and account details out of the repository.
