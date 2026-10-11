-- name: CreateJob :one
INSERT INTO jobs (
    external_id, job_source, title, company_name, description, source_url,
    salary_min, salary_max, currency, job_location, is_remote, skills, logo_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (external_id, job_source)
WHERE created_by_user_id IS NULL DO NOTHING
RETURNING *;

-- name: CreateImportedJob :one
INSERT INTO jobs (
    external_id, job_source, title, company_name, description, source_url,
    salary_min,
    salary_max,
    currency,
    job_location,
    is_remote,
    skills,
    logo_url,
    created_by_user_id, industry, category, company_stage, employment_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
ON CONFLICT (created_by_user_id, source_url)
WHERE created_by_user_id IS NOT NULL
DO UPDATE
    SET
        industry=excluded.industry, category=excluded.category, company_stage=excluded.company_stage, employment_type=excluded.employment_type,
        title = excluded.title,
        company_name = excluded.company_name,
        description = excluded.description,
        salary_min = excluded.salary_min,
        salary_max = excluded.salary_max,
        currency = excluded.currency,
        job_location = excluded.job_location,
        is_remote = excluded.is_remote,
        skills = excluded.skills,
        logo_url = excluded.logo_url,
        updated_at = NOW()
RETURNING *;

-- name: GetJobs :many
SELECT
    j.id,
    j.external_id,
    j.job_source,
    j.title,
    j.employment_type,
    j.industry, j.category, j.company_stage,
    j.company_name,
    j.description,
    j.source_url,
    j.salary_min,
    j.salary_max,
    j.currency,
    j.job_location,
    j.is_remote,
    j.skills,
    j.logo_url,
    j.created_by_user_id,
    j.posted_at,
    j.expires_at,
    j.created_at,
    j.updated_at,
    m.match_score,
    m.title_score,
    m.skill_score,
    m.experience_score,
    m.matched_skills,
    m.missing_skills,
    m.ai_summary,
    m.is_enriched,
    m.preference_eligible,
    m.preference_notes,
    m.created_at AS match_created_at,
    m.updated_at AS match_updated_at,
    (s.id IS NOT NULL)::BOOLEAN AS is_saved,
    (a.id IS NOT NULL)::BOOLEAN AS is_applied,
    (m.id IS NOT NULL)::BOOLEAN AS is_matched
FROM jobs AS j
LEFT JOIN saved_jobs AS s ON j.id = s.job_id AND s.user_id = $1
LEFT JOIN applications AS a ON j.id = a.job_id AND a.user_id = $1
LEFT JOIN user_job_matches AS m ON j.id = m.job_id AND m.user_id = $1
WHERE (j.created_by_user_id IS NULL OR j.created_by_user_id = $1)
AND (sqlc.arg(source)::text = '' OR lower(j.job_source) = sqlc.arg(source))
AND (sqlc.arg(search)::text = '' OR j.title ILIKE '%' || sqlc.arg(search) || '%' OR j.company_name ILIKE '%' || sqlc.arg(search) || '%' OR array_to_string(j.skills, ' ') ILIKE '%' || sqlc.arg(search) || '%')
AND (sqlc.arg(filter)::text = 'all' OR (sqlc.arg(filter)='saved' AND s.id IS NOT NULL) OR (sqlc.arg(filter)='applied' AND a.id IS NOT NULL)
 OR (sqlc.arg(filter)='remote' AND j.is_remote=true) OR (sqlc.arg(filter)='matched' AND m.match_score >= COALESCE((SELECT min_score FROM user_match_settings WHERE user_id=$1),60)) OR (sqlc.arg(filter)='today' AND j.created_at >= date_trunc('day',now())))
ORDER BY CASE WHEN sqlc.arg(sort_order)::text='score_desc' THEN m.match_score END DESC NULLS LAST,
 CASE WHEN sqlc.arg(sort_order)::text='score_asc' THEN m.match_score END ASC NULLS LAST,
 CASE WHEN sqlc.arg(sort_order)::text='oldest' THEN j.created_at END ASC,
 j.created_at DESC, j.id DESC
LIMIT $2 OFFSET $3;

-- name: GetAllJobIDs :many
SELECT id FROM jobs
WHERE created_by_user_id IS NULL OR created_by_user_id = $1;

-- name: GetJobByID :one
SELECT
    j.id,
    j.external_id,
    j.job_source,
    j.title,
    j.employment_type,
    j.industry, j.category, j.company_stage,
    j.company_name,
    j.description,
    j.source_url,
    j.salary_min,
    j.salary_max,
    j.currency,
    j.job_location,
    j.is_remote,
    j.skills,
    j.logo_url,
    j.created_by_user_id,
    j.created_at,
    j.updated_at,
    m.match_score,
    m.title_score,
    m.skill_score,
    m.experience_score,
    m.matched_skills,
    m.missing_skills,
    m.ai_summary,
    m.is_enriched,
    m.preference_eligible,
    m.preference_notes,
    m.created_at AS match_created_at,
    m.updated_at AS match_updated_at,
    (s.id IS NOT NULL)::BOOLEAN AS is_saved,
    (a.id IS NOT NULL)::BOOLEAN AS is_applied,
    (m.id IS NOT NULL)::BOOLEAN AS is_matched
FROM jobs AS j
LEFT JOIN saved_jobs AS s ON j.id = s.job_id AND s.user_id = $1
LEFT JOIN applications AS a ON j.id = a.job_id AND a.user_id = $1
LEFT JOIN user_job_matches AS m ON j.id = m.job_id AND m.user_id = $1
WHERE j.id = $2 AND (j.created_by_user_id IS NULL OR j.created_by_user_id = $1);

-- name: GetJobStats :one
SELECT
    COUNT(*) AS total_jobs,
    COUNT(*) FILTER (
        WHERE created_at >= DATE_TRUNC('day', NOW())
    ) AS new_jobs_today,
    MAX(created_at) AS latest_scrape_at,
    COALESCE(array_agg(DISTINCT job_source ORDER BY job_source), ARRAY[]::text[])::text[] AS sources
FROM jobs WHERE created_by_user_id IS NULL;

-- name: SearchJobs :many
SELECT
    id,
    external_id,
    job_source,
    title,
    company_name,
    description,
    source_url,
    salary_min,
    salary_max,
    currency,
    job_location,
    is_remote,
    skills,
    logo_url,
    created_at,
    updated_at
FROM jobs
WHERE
    (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
    AND ($2::TEXT IS NULL OR $2 = ANY(skills))
    AND (created_by_user_id IS NULL OR created_by_user_id = $3)
ORDER BY created_at DESC;

-- name: UpsertPublicJob :one
INSERT INTO jobs (external_id, job_source, title, company_name, description, source_url,
 salary_min, salary_max, currency, job_location, is_remote, skills, logo_url, posted_at, category, employment_type)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
ON CONFLICT (external_id, job_source) WHERE created_by_user_id IS NULL
DO UPDATE SET title=excluded.title, company_name=excluded.company_name,
 description=excluded.description, source_url=excluded.source_url,
 salary_min=excluded.salary_min, salary_max=excluded.salary_max, currency=excluded.currency,
 job_location=excluded.job_location, is_remote=excluded.is_remote, skills=excluded.skills,
 category=excluded.category, employment_type=excluded.employment_type,
 logo_url=excluded.logo_url, posted_at=excluded.posted_at, last_seen_at=now(),
 updated_at=CASE WHEN (jobs.title,jobs.company_name,jobs.description,jobs.source_url,
 jobs.salary_min,jobs.salary_max,jobs.currency,jobs.job_location,jobs.is_remote,jobs.skills,jobs.category,jobs.employment_type)
 IS DISTINCT FROM (excluded.title,excluded.company_name,excluded.description,excluded.source_url,
 excluded.salary_min,excluded.salary_max,excluded.currency,excluded.job_location,excluded.is_remote,excluded.skills,excluded.category,excluded.employment_type)
 THEN now() ELSE jobs.updated_at END
RETURNING *;

-- name: DeleteExpiredPublicJobs :execrows
-- Retention: delete public source postings whose age (posting date, else first-seen date)
-- is before the cutoff. Deleting a job cascades to applications, saved jobs, matches,
-- notifications and the digest ledger, so jobs the user saved or applied to are kept,
-- and private imports are never touched. Oldest first, in bounded batches so a large
-- backlog never holds long locks.
DELETE FROM jobs
WHERE jobs.id IN (
    SELECT j.id FROM jobs j
    WHERE j.created_by_user_id IS NULL
      AND COALESCE(j.posted_at, j.created_at) < sqlc.arg(cutoff)::timestamptz
      AND NOT EXISTS (SELECT 1 FROM saved_jobs s WHERE s.job_id = j.id)
      AND NOT EXISTS (SELECT 1 FROM applications a WHERE a.job_id = j.id)
    ORDER BY COALESCE(j.posted_at, j.created_at)
    LIMIT sqlc.arg(batch_size)::int
);
