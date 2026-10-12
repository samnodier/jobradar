# Reconciliation release review

This release is implemented on `feature/release-reconciliation` and has not been deployed. Review it locally before promoting to `main`, which triggers production deployment. Keep the existing Mac checkout and its private configuration; no folder deletion or history reset is needed.

## Local review

On every development machine, set `DIGEST_ENABLED=false` and `OPS_EMAIL_ENABLED=false` in the private `.env`. Stop an existing backend before migrating or starting another copy. Back up the development database first, then run:

```sh
make up
make migrate-up
make run
```

In another terminal, run `npm ci` and `npm run dev` from `frontend/`. New migrations add feedback, a separate Adzuna search location and link-check records.

1. **Login and profile:** sign in normally; check the saved resume, username and matching settings. An expired or spent password-setup link should offer normal sign-in with clear guidance. Never share a live setup link.
2. **Feedback:** open a job, choose Good match, reload and check it remains selected. Try Not relevant, then clear it. Wait for the matching backlog to drain. Feedback adjusts similar-job rankings without model calls; negative feedback excludes that specific job from new digest selection and sets its rematched score to zero. Clearing it permits normal rematching. Scores remain estimates, not hiring probabilities.
3. **Copy for AI:** try all six preparation modes: complete preparation, evaluation, tailored resume, cover letter, interview and outreach. Inspect copied text before pasting it into a chat. It should contain the selected job and saved profile, without provider keys or credentials. Copying sends no request to a model provider.
4. **Applications:** set a follow-up date in the past and check the due list; reschedule or clear it. Inspect Outcomes by source and score after changing a stage. The table summarizes current stages and current scores, not historical conversion rates.
5. **Link import and availability:** import a public posting, inspect the extracted fields and correct missing information before saving. Test manual entry if the page is blocked. A reachable page does not prove a vacancy remains open; unknown or blocked checks are distinct from confirmed unavailable links. The digest checks candidate links before delivery.
6. **Adzuna:** save a supported market and its separate search location, reload and verify persistence. For Johannesburg select ZA and enter Johannesburg; a blank search location searches the selected market. Keep your desired matching locations independently. A successful request returning zero jobs is different from a failed source request.
7. **System health:** check the running revision, process start, latest update outcome, local backup timestamp and alerts. Gray source-history bars mean no recorded attempts; percentages describe fetch attempts, not provider uptime. Refresh reads status and does not initiate collection. A local binary can report an unknown revision; verify the full release SHA on production.
8. **Digest preview:** run `make digest-preview` and review job eligibility, recipient, schedule and exclusions. It sends no email. Do not use `make digest-once` as a smoke test: it can send real mail.

Automated verification passed: the complete backend race suite with real isolated PostgreSQL/Redis, 61 frontend tests, type checking, production build and focused ESLint, six deployment stub tests, packaged image revision/SPA/non-root monitor checks, and new-migration down/up checks in an isolated database. Read-only public-feed checks returned 600 Jobicy, 99 Kenya and 100 South Africa records; this does not establish production ingestion or matching quality. Native PDF printing, real provider delivery and unattended operation still need manual verification.

## Promote after review

Review the feature commits, merge the verified branch into `dev`, then promote `dev` to `main` and push those branches using the agreed Git workflow. A `main` push starts CI and image publication. The server waits for both images tagged with the selected commit SHA before stopping the app, backing up, migrating and restarting. Confirm System health's running revision matches that SHA; a healthy response from an older release is insufficient.

## Activate on production

After the new image is running, edit only the server's private `.env` to enable the requested sources and operational email:

```dotenv
JOBICY_ENABLED=true
MYJOBMAG_ENABLED=true
MYJOBMAG_ZA_ENABLED=true
OPS_EMAIL_ENABLED=true
```

Keep `DIGEST_ENABLED=true` only on the single production sender, with the existing verified provider and owner recipient configuration. Development machines keep digest and operational email disabled. Changing `.env` alone does not update an existing container. From the production checkout, recreate the app using the exact running release SHA (replace the placeholder), then reinstall timers:

```sh
JOBRADAR_TAG=REVIEWED_FULL_COMMIT_SHA docker compose -p jobradar -f compose.production.yml up -d --force-recreate app
./deploy/install-systemd.sh
systemctl --user list-timers 'jobradar-*'
```

The installer adds the independent 15-minute monitor alongside update and backup timers. Check System health and `journalctl --user -u jobradar-monitor` for results. Enabled incident alerts can send owner email during normal monitor runs; no deliberate real-mail test is included here. Confirm the next expected digest and any actual alert in the inbox and provider history. No qualifying jobs can legitimately mean no digest. A whole-host, Docker, network or email-provider outage still needs monitoring outside that host.

After source activation, inspect their first recorded collection attempts and sample original links, attribution and geographic restrictions. Jobicy and MyJobMag poll at least six hours apart; Adzuna polls every twelve hours. Do not repeatedly restart to force requests. Read-only feed verification does not require AI keys.

GitHub login already exists. To activate it in production, configure private GitHub OAuth credentials and the correct HTTPS callback in both the app and GitHub. GitHub supports additional callback URLs; a separate production OAuth app is an option, not a requirement. Check actual consent and account linking before relying on it.

## Recovery copy and final acceptance

An off-server backup is a copy on another device, protecting your profile, saved jobs, applications and account if the server disk fails. Deleting old public jobs is separate housekeeping. On the Mac, use your configured SSH destination:

```sh
./deploy/copy-backup.sh YOUR_SSH_HOST "$HOME/jobradar-backups"
```

Keep the stable encryption key and production configuration separately and privately. The script copies the newest completed server dump; it does not restore it or delete server backups. Rehearse restoring into an empty isolated database and verify profile, applications and delivery ledger. Keep a small rotating set of recovery copies; avoid deleting your only verified backup.

Accept the release after local review, exact production revision verification, first successful enabled-source collection, scheduled digest confirmation, backup restoration practice and a week of unattended operation. Remaining product research includes historical outcome calibration, richer ATS sources and external outage monitoring; the current outcome table and feedback adjustment do not yet provide those guarantees.
