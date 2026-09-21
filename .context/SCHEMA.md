## Database Schema

**users** — `id`, `email`, `username`, `full_name`, `avatar_url`, `phone`, `user_location`, `website_url`, `linkedin_url`, `github_url`, `headline`, `user_summary`, `availability`, `min_salary`, `max_salary`, `salary_currency`, `years_of_experience`, `preferred_job_types`, `preferred_industries`, `company_stage_preference`, `notify_jobs`, `is_admin`, `created_at`, `updated_at`
*(`encrypted_gemini_api_key` was dropped in migration `20260706100000` — keys now live in `user_api_keys`)*

**user_api_keys** — `id`, `user_id` (FK→users CASCADE), `provider` CHECK('gemini','groq'), `encrypted_key` (AES-256-GCM ciphertext; never sent to client), `created_at`, `updated_at`, UNIQUE(user_id, provider)

**user_accounts** — `id`, `user_id` (FK→users CASCADE), `auth_provider`, `auth_provider_id`, `access_token`, `created_at`

**jobs** — `id`, `external_id`, `job_source`, `title`, `company_name`, `description`, `source_url`, `salary_min`, `salary_max`, `currency`, `job_location`, `is_remote`, `skills TEXT[]`, `logo_url`, `created_at`, `updated_at`

**saved_jobs** — `id`, `user_id` (FK→users CASCADE), `job_id` (FK→jobs CASCADE), `created_at`, UNIQUE(user_id, job_id)

**applications** — `id`, `user_id` (FK→users CASCADE), `job_id` (FK→jobs CASCADE), `status` CHECK('applied','recruiter_screen','interview','offer','accepted','rejected','withdrawn'), `applied_at`, `last_status_changed_at`, `follow_up_at`, `notes`, `created_at`, `updated_at`, UNIQUE(user_id, job_id)

**user_job_matches** — `id`, `user_id` (FK→users CASCADE), `job_id` (FK→jobs CASCADE), `match_score INTEGER`, `title_score FLOAT`, `skill_score FLOAT`, `experience_score FLOAT`, `matched_skills TEXT[]`, `missing_skills TEXT[]`, `ai_summary TEXT`, `is_enriched BOOLEAN DEFAULT FALSE`, `created_at`, `updated_at`, UNIQUE(user_id, job_id)

**user_skills** — `id`, `user_id` (FK→users CASCADE), `skill_name`, `proficiency`, `years_experience`, `created_at`

**user_desired_roles** — `id`, `user_id` (FK→users CASCADE), `role_title`, `created_at`

**user_desired_locations** — `id`, `user_id` (FK→users CASCADE), `location`, `created_at`

*(Resume profile tables: user_experiences, user_education, user_projects, user_certifications, user_languages — see resume_sections migration)*

## SQLC Queries Written

**jobs:** `CreateJob`, `GetJobs` (LEFT JOINs saved_jobs + applications + user_job_matches), `GetJobByID`, `GetJobStats`, `SearchJobs`

**user_job_matches:** `UpsertMatch`, `UpdateMatchEnrichment` (narrow: sets `ai_summary` + `is_enriched = TRUE` only, scoped by user_id AND job_id)

**saved_jobs:** `GetSavedJobsForUser`, `SaveJob`, `UnSaveJob`

**applications:** `CreateApplication`, `GetApplicationByID`, `GetApplicationByUserAndJob`, `GetApplicationsByUserID`, `UpdateApplicationStatus`, `UpdateApplicationNotes`, `UpdateApplicationFollowUp`, `DeleteApplication`

**users:** `CreateUser`, `GetUserByID` (derives `configured_providers text[]` via ARRAY subquery on user_api_keys; never selects ciphertext), `GetUserByProviderIdentity`, `UpdateUser`, `DeleteUserByID`, `GetAllUsers`

**user_api_keys:** `UpsertUserAPIKey` (ON CONFLICT (user_id, provider) replaces the ciphertext), `GetUserAPIKey` (read ciphertext for workers/handlers only), `ListUserAPIKeyProviders`, `DeleteUserAPIKey`

## September 2026 corrections and additions

The original abbreviated list above is historical; migration files and SQLC models are authoritative.

- jobs also includes `created_by_user_id` (nullable owner FK, cascades), `job_status`, `employment_type`, `experience_level`, `posted_at`, `expires_at`, and `last_seen_at`. Public uniqueness is `(external_id,job_source)` WHERE owner IS NULL; private uniqueness is `(created_by_user_id,source_url)` WHERE owner IS NOT NULL.
- user_skills uses `(user_id,skill_id)` as its primary key; skill names belong in the case-insensitive skills table.
- applications uses `application_status`; updates use the consolidated `UpdateApplication` query.
- pending_matches: `(user_id,job_id)` PK with cascading FKs, `revision BIGINT`, nullable `enqueued_at`. Job/profile triggers schedule work transactionally. Dispatcher retries delivery after five minutes; completion matches the revision.
- email_digests: id, user_id, digest_date, recipient, sender, provider, subject, html, status, attempts, first_attempt_at, next_attempt_at, provider_message_id, last_error, created_at, sent_at. UNIQUE(user_id,digest_date). States: pending, sending, sent, uncertain, failed.
- email_digest_jobs: digest_id, user_id, job_id; PRIMARY KEY(user_id,job_id), cascading FKs. Reserves jobs in the same transaction as digest preparation.
- New queries: UpsertPublicJob; ListPendingMatches, MarkMatchEnqueued, CompletePendingMatch, GetPendingMatchRevision, ScheduleUserMatches; digest preparation/delivery/history queries; ListServiceHealth.
- GetJobs now takes bounded pagination, search, filter, and sort parameters. SaveJob and CreateApplication insert only from a public-or-owned job selection.

## Profile workflows

- user_match_settings: user_id PK/FK, min_score 0–100, work_mode any/remote/onsite, include_unknown, search_country, email_enabled, notifications_enabled, digest_timezone, digest_hours (1–2 hours 0–23), digest_max_jobs 1–25. GET supplies defaults when a row is absent.
- user_job_matches adds preference_eligible and preference_notes; ranking and eligibility are separate.
- email_digests adds digest_slot, changing uniqueness to (user_id,digest_date,digest_slot). User/job reservations still prevent repeat job delivery across slots/days.
- match_notifications: id, user_id, job_id, created_at, read_at; UNIQUE(user_id,job_id); ownership FKs cascade. List joins current eligible scores/settings.
- Settings, desired-location and relevant users preference changes schedule rematching.

## Preference ranking and read tokens

- jobs adds nullable industry, category and company_stage. Adzuna category/contract time and reviewed import metadata are persisted. Industry/company-stage profile changes trigger rematching.
- read_api_tokens: UUID id, user_id FK cascade, name, unique token_hash, created_at, expires_at, revoked_at. Only token hashes are stored.
- There is no local password hash or email/password login yet.
