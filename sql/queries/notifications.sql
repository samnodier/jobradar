-- name: CreateMatchNotification :exec
INSERT INTO match_notifications(user_id,job_id)VALUES($1,$2) ON CONFLICT(user_id,job_id) DO NOTHING;

-- name: ListMatchNotifications :many
SELECT n.id,n.job_id,n.created_at,n.read_at,j.title,j.company_name,m.match_score
FROM match_notifications n JOIN jobs j ON j.id=n.job_id
JOIN user_job_matches m ON m.user_id=n.user_id AND m.job_id=n.job_id
JOIN users u ON u.id=n.user_id LEFT JOIN user_match_settings s ON s.user_id=n.user_id
WHERE n.user_id=$1 AND (j.created_by_user_id IS NULL OR j.created_by_user_id=$1)
AND u.notify_jobs=true AND COALESCE(s.notifications_enabled,true)=true
AND m.preference_eligible=true AND m.match_score>=COALESCE(s.min_score,60)
AND NOT EXISTS(SELECT 1 FROM pending_matches p WHERE p.user_id=n.user_id AND p.job_id=n.job_id)
ORDER BY n.created_at DESC,n.id LIMIT 50;

-- name: ReadMatchNotification :execrows
UPDATE match_notifications SET read_at=COALESCE(read_at,now()) WHERE user_id=$1 AND id=$2;
