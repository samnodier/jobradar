## Current routes

All `/api` responses are non-cacheable. Job reads allow anonymous access to public jobs and add the signed-in user's own private jobs and match data. All mutations and profile/digest/system endpoints require a Redis session. Export routes additionally accept revocable read-only bearer tokens. PERSONAL_EMAIL can restrict login to the owner.

| Method | Route | Access / behavior |
| --- | --- | --- |
| GET | `/api/health` | Public liveness |
| GET | `/api/jobs` | Public-or-owned; limit 1–500 (default 50), offset, q, source (case-normalized exact match before pagination), filter=all/saved/applied/remote/today/matched (matched uses only the saved score threshold; source/search still apply), order=newest/oldest/score_desc/score_asc (default score_desc for matched, newest otherwise); array response with X-Has-More header |
| GET | `/api/jobs/stats` | Public-job totals, added-today count, latest addition and actual source list |
| GET | `/api/jobs/{jobID}` | Public-or-owned details and per-user match |
| POST | `/api/jobs/import` | Auth; extract draft from URL; no database write |
| POST | `/api/jobs/import/confirm` | Auth; validate/upsert private job and atomically schedule matching |
| GET/POST/DELETE | `/api/saved_jobs` | Auth; own saves; POST/DELETE body job_id |
| GET/POST | `/api/applications` | Auth; list/create |
| GET | `/api/applications/{applicationID}` | Auth; own application |
| PATCH/DELETE | `/api/applications/{id}` | Auth; own application |
| GET/PATCH/DELETE | `/api/users/me` | Auth; own account |
| PUT/DELETE | `/api/users/me/api-keys/{provider}` | Auth; gemini or groq, encrypted at rest |
| GET/POST | `/api/profile/experiences`, `/educations`, `/projects`, `/certifications`, `/languages` | Auth; own profile records (each collection is under /api/profile) |
| PATCH/DELETE | `/api/profile/{collection}/{id}` | Auth; own record; language uses its language identifier |
| GET/PATCH | `/api/profile/preferences` | Auth; own preferences; triggers rematching for relevant rows |
| GET/PUT | `/api/profile/matching` | Auth; own score/filter/search-market/delivery settings; PUT validates complete resource |
| GET | `/api/jobs/{jobID}/ai-context` | Auth; public-or-owned job plus own resume, Markdown, no AI call |
| GET | `/api/notifications` | Auth; latest 50 currently eligible own match notifications |
| PATCH | `/api/notifications/{notificationID}/read` | Auth; mark own notification read |
| GET | `/api/digest` | Auth; own digest status/history, no send |
| GET | `/api/digest/preview` | Auth; configured personal recipient only; HTML preview, no send |
| GET | `/api/system/status` | Auth; DB/Redis connectivity, backlog and source health with last_checked and 30 UTC-day success/failure counts; history_available flags partial history failure |
| GET | `/auth/github/login` | Optional OAuth redirect; link=1 requires a recent session |
| GET | `/auth/github/callback` | OAuth state-validated callback |
| GET | `/auth/users/me` | Auth user for frontend boot |
| POST | `/auth/logout` | Clear session |
| GET/POST | `/api/auth/onboarding` | Pending-signup token |
| GET | Other non-API browser routes | Built Vue SPA history fallback |

See `cmd/jobradar/main.go` and `internal/auth/handler.go` for canonical route registration.

## Read-only workflow API

- GET `/api/export/jobs`: session or read token; score/date/timezone/source/location filters and score/UUID keyset pagination. Defaults to fresh eligible matches at the profile threshold. See `docs/JOB_EXPORT_API.md`.
- GET `/api/export/jobs/{jobID}/ai-context`: same read authentication, public-or-owned job and own saved resume Markdown.
- GET/POST `/api/profile/access-tokens`, DELETE `/api/profile/access-tokens/{tokenID}`: browser session only; owner-scoped creation/list/revocation.

## Local/password authentication

- GET `/auth/methods`: public enabled login methods, no secrets.
- POST `/auth/password/login`: email/password, rate-limited, JSON/origin validation, Redis session.
- POST `/auth/password/setup`: private server-issued token plus chosen password; single-use, no public registration.
- GET `/auth/password`: session; own email, whether password is configured, GitHub server availability and account-specific github_connected state.
- PUT `/auth/password`: session; add after recent GitHub login or change with current password; invalidates other sessions, read tokens and setup links.
- Server CLI `--password-setup EMAIL`: provision/recover a local account and print a private 15-minute setup link. No password argument; preserves an existing user ID.

The public homepage requests the newest public jobs (order=newest) and omits session credentials for that list.

- GET/PUT/DELETE `/api/jobs/{jobID}/feedback`: authenticated own vote for a public-or-owned job; PUT accepts sentiment -1 or 1; DELETE clears and schedules rematching.
- AI-context routes accept optional `mode=evaluate|resume|cover-letter|interview|outreach`; omitted mode retains combined preparation. Invalid modes return 400.
- System status includes the running binary revision and process start time; public health includes revision for deployment verification.

- GET/POST `/api/jobs/{jobID}/availability`: authenticated public-or-owned cached availability; POST performs a SSRF-protected link check at most once daily. Reachable does not mean confirmed open; 404/410 is unavailable; network/access failures are unknown.
