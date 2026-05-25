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
    last_status_changed_at,
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
    last_status_changed_at,
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
    a.last_status_changed_at,
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

-- name: UpdateApplication :one
UPDATE applications
SET
    notes = COALESCE(sqlc.narg('notes'), notes),
    application_status = COALESCE(
        sqlc.narg('application_status'), application_status
    ),
    applied_at = CASE
        WHEN sqlc.narg('clear_applied_at')::boolean = true THEN null
        WHEN
            sqlc.narg('applied_at')::timestamptz IS NOT null
            THEN sqlc.narg('applied_at')::timestamptz
        ELSE applied_at
    END,
    follow_up_at = CASE
        WHEN sqlc.narg('clear_follow_up_at')::boolean = true THEN null
        WHEN
            sqlc.narg('follow_up_at')::timestamptz IS NOT null
            THEN sqlc.narg('follow_up_at')::timestamptz
        ELSE follow_up_at
    END,
    last_status_changed_at = CASE
        WHEN
            sqlc.narg('application_status') IS NOT null
            AND sqlc.narg('application_status') <> application_status THEN NOW()
        ELSE last_status_changed_at
    END,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteApplication :exec
DELETE FROM applications
WHERE id = $1 AND user_id = $2;
