-- name: GetSavedJobsForUser :many
SELECT s.id, j.id as job_id, j.title, j.company, j.location, j.is_remote, j.logo_url, a.status, s.created_at AS saved_at
FROM saved_jobs s
JOIN jobs j ON s.job_id = j.id
LEFT JOIN applications a ON s.job_id = a.job_id AND s.user_id = a.user_id
WHERE s.user_id = $1
ORDER BY s.created_at DESC;

-- name: SaveJob :one
INSERT INTO saved_jobs (user_id, job_id)
VALUES ($1, $2)
RETURNING *;

-- name: UnSaveJob :exec
DELETE FROM saved_jobs
WHERE id = $1 AND user_id = $2;