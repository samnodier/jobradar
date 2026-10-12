-- name: SetJobFeedback :one
INSERT INTO job_feedback(user_id,job_id,sentiment)
SELECT sqlc.arg(user_id), j.id, sqlc.arg(sentiment)::smallint FROM jobs j
WHERE j.id=sqlc.arg(job_id) AND (j.created_by_user_id IS NULL OR j.created_by_user_id=sqlc.arg(user_id))
ON CONFLICT(user_id,job_id) DO UPDATE SET sentiment=excluded.sentiment, updated_at=now()
RETURNING sentiment;

-- name: ClearJobFeedback :exec
DELETE FROM job_feedback WHERE user_id=$1 AND job_id=$2;

-- name: GetJobFeedback :one
SELECT sentiment FROM job_feedback WHERE user_id=$1 AND job_id=$2;

-- name: ListJobFeedback :many
SELECT j.id, j.title, j.skills, f.sentiment FROM job_feedback f JOIN jobs j ON j.id=f.job_id
WHERE f.user_id=$1 ORDER BY f.updated_at DESC LIMIT 200;
