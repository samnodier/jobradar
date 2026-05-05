-- name: CreateApplication :one
INSERT INTO applications (
    user_id, job_id, status, applied_at, notes, follow_up_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetApplicationByID :one
SELECT * FROM applications
WHERE id = $1 AND user_id = $2;

-- name: GetApplicationByUserAndJob :one
SELECT * FROM applications
WHERE user_id = $1 AND job_id = $2;

-- name: GetApplicationsByUserID :many
SELECT 
    a.*,
    j.title AS job_title,
    j.company AS job_company,
    j.location AS job_location,
    j.url AS job_url,
    j.is_remote AS job_is_remote,
    j.logo_url AS job_logo_url
FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE a.user_id = $1
ORDER BY a.updated_at DESC;

-- name: UpdateApplicationStatus :one
UPDATE applications
SET 
    status = $2,
    last_status_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: UpdateApplicationNotes :one
UPDATE applications
SET 
    notes = $2,
    updated_at = NOW()
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: UpdateApplicationFollowUp :one
UPDATE applications
SET 
    follow_up_at = $2,
    updated_at = NOW()
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: DeleteApplication :exec
DELETE FROM applications
WHERE id = $1 AND user_id = $2;
