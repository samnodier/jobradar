-- name: CreateJob :one
INSERT INTO jobs (
    external_id, source, title, company, description, url,
    salary_min, salary_max, currency, location, is_remote, skills, logo_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (external_id, source) DO NOTHING
RETURNING *;

-- name: GetJobs :many
SELECT * FROM jobs
ORDER BY created_at DESC;

-- name: GetJobByID :one
SELECT * FROM jobs WHERE id = $1;

-- name: SearchJobs :many
SELECT * FROM jobs
WHERE
    (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
    AND ($2::TEXT IS NULL OR $2 = ANY(skills))
ORDER BY created_at DESC;
