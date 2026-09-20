-- name: GetSavedJobsForUser :many
SELECT
    s.id,
    j.id AS job_id,
    j.title,
    j.company_name,
    j.job_location,
    j.is_remote,
    j.logo_url,
    a.application_status,
    s.created_at AS saved_at
FROM saved_jobs AS s
INNER JOIN jobs AS j ON s.job_id = j.id
LEFT JOIN applications AS a ON s.job_id = a.job_id AND s.user_id = a.user_id
WHERE s.user_id = $1 AND (j.created_by_user_id IS NULL OR j.created_by_user_id = $1)
ORDER BY s.created_at DESC;

-- name: SaveJob :one
INSERT INTO saved_jobs (user_id, job_id)
SELECT $1, j.id FROM jobs j
WHERE j.id = $2 AND (j.created_by_user_id IS NULL OR j.created_by_user_id = $1)
RETURNING *;

-- name: UnSaveJob :exec
DELETE FROM saved_jobs
WHERE job_id = $1 AND user_id = $2;
