-- name: UpsertMatch :exec
INSERT INTO user_job_matches (
    user_id,
    job_id,
    match_score,
    title_score,
    skill_score,
    experience_score,
    matched_skills,
    missing_skills,
    ai_summary,
    is_enriched
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (user_id, job_id) DO UPDATE SET
    match_score = excluded.match_score,
    title_score = excluded.title_score,
    skill_score = excluded.skill_score,
    experience_score = excluded.experience_score,
    matched_skills = excluded.matched_skills,
    missing_skills = excluded.missing_skills,
    ai_summary = excluded.ai_summary,
    is_enriched = excluded.is_enriched,
    updated_at = NOW();

-- name: UpdateMatchEnrichment :exec
UPDATE user_job_matches
SET
    ai_summary = $3,
    is_enriched = TRUE,
    updated_at = NOW()
WHERE user_id = $1 AND job_id = $2;
