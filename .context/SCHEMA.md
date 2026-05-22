## Database Schema

**users** — `id`, `email`, `username`, `name`, `avatar_url`, `created_at`, `updated_at`

**user_accounts** — `id`, `user_id` (FK→users CASCADE), `provider`, `provider_id`, `access_token`, `created_at`

**jobs** — `id`, `external_id`, `source`, `title`, `company`, `location`, `is_remote`, `url`, `salary`, `logo_url`, `posted_at`, `raw_description`, `match_score`, `created_at`

**saved_jobs** — `id`, `user_id` (FK→users CASCADE), `job_id` (FK→jobs CASCADE), `created_at` NOT NULL, UNIQUE(user_id, job_id)

**applications** — `id`, `user_id` (FK→users CASCADE), `job_id` (FK→jobs CASCADE), `status` CHECK('applied','recruiter_screen','interview','offer','accepted','rejected','withdrawn'), `applied_at`, `last_status_changed_at`, `follow_up_at`, `notes`, `created_at`, `updated_at`, UNIQUE(user_id, job_id)

## SQLC Queries Written

**saved_jobs:** `GetSavedJobsForUser`, `SaveJob`, `UnSaveJob`

**applications:** `CreateApplication`, `GetApplicationByID`, `GetApplicationByUserAndJob`, `GetApplicationsByUserID`, `UpdateApplicationStatus`, `UpdateApplicationNotes`, `UpdateApplicationFollowUp`, `DeleteApplication`

**users:** `CreateUser`, `GetUserByID`, `GetUserByProviderIdentity`, `DeleteUserByID`
