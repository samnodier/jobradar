-- name: GetDigestUser :one
SELECT id FROM users WHERE email = $1 AND notify_jobs = true;

-- name: GetDigestCandidates :many
SELECT j.id, j.title, j.company_name, j.source_url, j.job_source, j.job_location,
       j.currency, j.salary_min, j.salary_max, m.match_score
FROM jobs j JOIN user_job_matches m ON m.job_id=j.id AND m.user_id=$1
WHERE (j.created_by_user_id IS NULL OR j.created_by_user_id=$1)
 AND m.match_score >= $2
 AND COALESCE(j.posted_at,j.created_at) >= now() - interval '30 days'
 AND (j.expires_at IS NULL OR j.expires_at > now())
 AND NOT EXISTS (SELECT 1 FROM pending_matches p WHERE p.user_id=$1 AND p.job_id=j.id)
 AND NOT EXISTS (SELECT 1 FROM email_digest_jobs d WHERE d.user_id=$1 AND d.job_id=j.id)
ORDER BY m.match_score DESC, j.created_at DESC, j.id
LIMIT $3;

-- name: CreateDigest :one
INSERT INTO email_digests (user_id,digest_date,recipient,sender,provider,subject,html)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id,digest_date) DO NOTHING RETURNING *;

-- name: ReserveDigestJob :exec
INSERT INTO email_digest_jobs (digest_id,user_id,job_id) VALUES ($1,$2,$3);

-- name: GetDailyDigest :one
SELECT * FROM email_digests WHERE user_id=$1 AND digest_date=$2;

-- name: GetPendingDigest :one
SELECT * FROM email_digests WHERE recipient=$1 AND status IN ('pending','sending')
AND next_attempt_at <= now() ORDER BY created_at LIMIT 1;

-- name: StartDigestAttempt :exec
UPDATE email_digests SET status='sending', attempts=attempts+1,
first_attempt_at=COALESCE(first_attempt_at,now()),next_attempt_at=now()+interval '2 minutes'
WHERE id=$1;

-- name: FinishDigest :exec
UPDATE email_digests SET status='sent',sent_at=now(),provider_message_id=$2,last_error=NULL WHERE id=$1;

-- name: FailDigest :exec
UPDATE email_digests SET status=$2,last_error=$3,next_attempt_at=now()+interval '5 minutes' WHERE id=$1;

-- name: ListDigestHistory :many
SELECT id,digest_date,status,attempts,created_at,sent_at,last_error
FROM email_digests WHERE user_id=$1 ORDER BY created_at DESC LIMIT 10;
