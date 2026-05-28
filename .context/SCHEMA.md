## Database Schema

**users** — `id`, `email`, `username`, `full_name`, `avatar_url`, `phone`, `user_location`, `website_url`, `linkedin_url`, `github_url`, `headline`, `user_summary`, `availability`, `min_salary`, `max_salary`, `salary_currency`, `years_of_experience`, `preferred_job_types`, `preferred_industries`, `company_stage_preference`, `notify_jobs`, `is_admin`, `created_at`, `updated_at`

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

**user_job_matches:** `UpsertMatch`

**saved_jobs:** `GetSavedJobsForUser`, `SaveJob`, `UnSaveJob`

**applications:** `CreateApplication`, `GetApplicationByID`, `GetApplicationByUserAndJob`, `GetApplicationsByUserID`, `UpdateApplicationStatus`, `UpdateApplicationNotes`, `UpdateApplicationFollowUp`, `DeleteApplication`

**users:** `CreateUser`, `GetUserByID`, `GetUserByProviderIdentity`, `UpdateUser`, `DeleteUserByID`, `GetAllUsers`
