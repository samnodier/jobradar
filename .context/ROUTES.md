## Current routes

All `/api` responses are non-cacheable. Job reads allow anonymous access to public jobs and add the signed-in user's own private jobs and match data. All mutations and profile/digest/system endpoints require a Redis session. PERSONAL_EMAIL can restrict login to the owner.

| Method | Route | Access / behavior |
| --- | --- | --- |
| GET | `/api/health` | Public liveness |
| GET | `/api/jobs` | Public-or-owned; limit 1–100 (default 50), offset, q, filter=all/saved/applied/remote/today, order=newest/oldest |
| GET | `/api/jobs/stats` | Public-job totals |
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
| GET | `/api/digest` | Auth; own digest status/history, no send |
| GET | `/api/digest/preview` | Auth; configured personal recipient only; HTML preview, no send |
| GET | `/api/system/status` | Auth; DB/Redis connectivity, backlog and source health |
| GET | `/auth/github/login` | Public OAuth redirect |
| GET | `/auth/github/callback` | OAuth state-validated callback |
| GET | `/auth/users/me` | Auth user for frontend boot |
| POST | `/auth/logout` | Clear session |
| GET/POST | `/api/auth/onboarding` | Pending-signup token |
| GET | Other non-API browser routes | Built Vue SPA history fallback |

See `cmd/jobradar/main.go` and `internal/auth/handler.go` for canonical route registration.
