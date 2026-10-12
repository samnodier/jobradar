-- name: GetMatchSettings :one
SELECT u.id AS user_id, COALESCE(s.min_score,60)::int AS min_score,
 COALESCE(s.work_mode,'any')::text AS work_mode, COALESCE(s.include_unknown,true)::boolean AS include_unknown,
 COALESCE(s.search_country,'us')::text AS search_country,
 COALESCE(s.search_location,'')::text AS search_location,
 COALESCE(s.email_enabled,true)::boolean AS email_enabled,
 COALESCE(s.notifications_enabled,true)::boolean AS notifications_enabled,
 COALESCE(s.digest_timezone,'America/New_York')::text AS digest_timezone,
 COALESCE(s.digest_hours,ARRAY[8,18])::int[] AS digest_hours,
 COALESCE(s.digest_max_jobs,10)::int AS digest_max_jobs
FROM users u LEFT JOIN user_match_settings s ON s.user_id=u.id WHERE u.id=$1;

-- name: UpsertMatchSettings :exec
INSERT INTO user_match_settings(user_id,min_score,work_mode,include_unknown,search_country,email_enabled,notifications_enabled,digest_timezone,digest_hours,digest_max_jobs,search_location)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT(user_id) DO UPDATE SET min_score=excluded.min_score,work_mode=excluded.work_mode,include_unknown=excluded.include_unknown,
search_country=excluded.search_country,email_enabled=excluded.email_enabled,notifications_enabled=excluded.notifications_enabled,
digest_timezone=excluded.digest_timezone,digest_hours=excluded.digest_hours,digest_max_jobs=excluded.digest_max_jobs,search_location=excluded.search_location;
