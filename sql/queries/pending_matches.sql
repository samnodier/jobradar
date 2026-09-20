-- name: ListPendingMatches :many
SELECT user_id, job_id, revision FROM pending_matches
WHERE enqueued_at IS NULL OR enqueued_at < now() - interval '5 minutes'
ORDER BY enqueued_at NULLS FIRST, user_id, job_id LIMIT 200;

-- name: MarkMatchEnqueued :exec
UPDATE pending_matches SET enqueued_at = now()
WHERE user_id = $1 AND job_id = $2 AND revision = $3;

-- name: CompletePendingMatch :exec
DELETE FROM pending_matches WHERE user_id = $1 AND job_id = $2 AND revision = $3;

-- name: GetPendingMatchRevision :one
SELECT revision FROM pending_matches WHERE user_id = $1 AND job_id = $2;

-- name: ScheduleUserMatches :exec
SELECT schedule_user_matches($1::uuid);
