-- name: CreateJob :one
INSERT INTO jobs (
    external_id, job_source, title, company_name, description, source_url,
    salary_min, salary_max, currency, job_location, is_remote, skills, logo_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (external_id, job_source) DO NOTHING
RETURNING *;

-- name: GetJobs :many
SELECT
    j.id,
    j.external_id,
    j.job_source,
    j.title,
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
    j.created_at,
    j.updated_at,
    (s.id IS NOT NULL)::BOOLEAN AS is_saved,
    (a.id IS NOT NULL)::BOOLEAN AS is_applied
FROM jobs AS j
LEFT JOIN saved_jobs AS s ON j.id = s.job_id AND s.user_id = $1
LEFT JOIN applications AS a ON j.id = a.job_id AND a.user_id = $1
ORDER BY j.created_at DESC;

-- name: GetJobByID :one
SELECT
    j.id,
    j.external_id,
    j.job_source,
    j.title,
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
    j.created_at,
    j.updated_at,
    (s.id IS NOT NULL)::BOOLEAN AS is_saved,
    (a.id IS NOT NULL)::BOOLEAN AS is_applied
FROM jobs AS j
LEFT JOIN saved_jobs AS s ON j.id = s.job_id AND s.user_id = $1
LEFT JOIN applications AS a ON j.id = a.job_id AND a.user_id = $1
WHERE j.id = $2;

-- name: GetJobStats :one
SELECT
    COUNT(*) AS total_jobs,
    COUNT(*) FILTER (
        WHERE created_at >= DATE_TRUNC('day', NOW())
    ) AS new_jobs_today,
    MAX(created_at) AS latest_scrape_at
FROM jobs;

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
ORDER BY created_at DESC;
