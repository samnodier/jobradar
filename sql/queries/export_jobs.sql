-- name: ExportJobs :many
SELECT
    j.id,
    j.external_id,
    j.job_source,
    j.title,
    j.employment_type,
    j.industry, j.category, j.company_stage,
    j.company_name,
    j.description,
    j.source_url,
    j.salary_min,
    j.salary_max,
    j.currency,
    j.job_location,
    j.is_remote,
    j.skills,
    j.logo_url,
    j.created_by_user_id,
    j.posted_at,
    j.expires_at,
    j.created_at,
    j.updated_at,
    m.match_score,
    m.title_score,
    m.skill_score,
    m.experience_score,
    m.matched_skills,
    m.missing_skills,
    m.ai_summary,
    m.is_enriched,
    m.preference_eligible,
    m.preference_notes,
    m.created_at AS match_created_at,
    m.updated_at AS match_updated_at,
    (s.id IS NOT NULL)::BOOLEAN AS is_saved,
    (a.id IS NOT NULL)::BOOLEAN AS is_applied,
    (m.id IS NOT NULL)::BOOLEAN AS is_matched
FROM jobs AS j
LEFT JOIN saved_jobs AS s ON j.id = s.job_id AND s.user_id = sqlc.arg(user_id)
LEFT JOIN applications AS a ON j.id = a.job_id AND a.user_id = sqlc.arg(user_id)
LEFT JOIN user_job_matches AS m ON j.id = m.job_id AND m.user_id = sqlc.arg(user_id)
WHERE (j.created_by_user_id IS NULL OR j.created_by_user_id=sqlc.arg(user_id))
AND m.match_score >= sqlc.arg(min_score)::integer
AND (NOT sqlc.arg(eligible_only)::boolean OR m.preference_eligible)
AND NOT EXISTS(SELECT 1 FROM pending_matches p WHERE p.user_id=sqlc.arg(user_id) AND p.job_id=j.id)
AND (sqlc.arg(source)::text='' OR j.job_source=sqlc.arg(source))
AND (NOT sqlc.arg(exclude_applied)::boolean OR a.id IS NULL)
AND (NOT sqlc.arg(remote_only)::boolean OR j.is_remote=true)
AND (sqlc.arg(include_expired)::boolean OR j.expires_at IS NULL OR j.expires_at>now())
AND (sqlc.arg(location)::text='' OR j.job_location ILIKE '%' || sqlc.arg(location) || '%')
AND (sqlc.narg(start_at)::timestamptz IS NULL OR
 CASE sqlc.arg(date_field)::text WHEN 'posted' THEN j.posted_at WHEN 'matched' THEN m.created_at ELSE j.created_at END >= sqlc.narg(start_at))
AND (sqlc.narg(end_at)::timestamptz IS NULL OR
 CASE sqlc.arg(date_field)::text WHEN 'posted' THEN j.posted_at WHEN 'matched' THEN m.created_at ELSE j.created_at END < sqlc.narg(end_at))
AND (m.match_score < sqlc.arg(before_score)::integer OR (m.match_score=sqlc.arg(before_score) AND j.id<sqlc.arg(before_id)::uuid))
ORDER BY m.match_score DESC,j.id DESC
LIMIT sqlc.arg(page_limit);
