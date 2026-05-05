-- name: CreateApplication :one
INSERT INTO applications (
    user_id, job_id, application_status, applied_at, notes, follow_up_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetApplicationByID :one
SELECT
    id,
    user_id,
    job_id,
    application_status,
    applied_at,
    notes,
    follow_up_at,
    created_at,
    updated_at
FROM applications
WHERE id = $1 AND user_id = $2;

-- name: GetApplicationByUserAndJob :one
SELECT
    id,
    user_id,
    job_id,
    application_status,
    applied_at,
    notes,
    follow_up_at,
    created_at,
    updated_at
FROM applications
WHERE user_id = $1 AND job_id = $2;

-- name: GetApplicationsByUserID :many
SELECT
    a.id,
    a.user_id,
    a.job_id,
    a.application_status,
    a.applied_at,
    a.notes,
    a.follow_up_at,
    a.created_at,
    a.updated_at,
    j.title AS job_title,
    j.company_name,
    j.job_location,
    j.source_url,
    j.is_remote AS job_is_remote,
    j.logo_url AS job_logo_url
FROM applications AS a
INNER JOIN jobs AS j ON a.job_id = j.id
WHERE a.user_id = $1
ORDER BY a.updated_at DESC;

-- name: UpdateApplicationStatus :one
UPDATE applications
SET
    application_status = $2,
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
