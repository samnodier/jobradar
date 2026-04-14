-- name: CreateJob :one
INSERT INTO jobs (
    external_id, source, title, company, description, url,
    salary_min, salary_max, currency, location, is_remote, skills, logo_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (external_id, source) DO NOTHING
RETURNING *;
