-- name: CreateUserExperience :one
INSERT INTO user_experiences (
    user_id,
    company_name,
    company_url,
    role_title,
    exp_location,
    industry,
    employment_type,
    description,
    achievements,
    start_date,
    end_date,
    is_current
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetExperiencesByUserID :many
SELECT
    e.id,
    e.user_id,
    e.company_name,
    e.company_url,
    e.role_title,
    e.exp_location,
    e.industry,
    e.employment_type,
    e.description,
    e.achievements,
    e.start_date,
    e.end_date,
    e.is_current,
    e.created_at,
    e.updated_at,
    COALESCE(
        JSON_AGG(
            JSON_BUILD_OBJECT('id', s.id, 'name', s.name)
        ) FILTER (WHERE s.id IS NOT NULL), '[]'
    ) AS skills
FROM user_experiences AS e
LEFT JOIN experience_skills AS es ON e.id = es.experience_id
LEFT JOIN skills AS s ON es.skill_id = s.id
WHERE e.user_id = $1
GROUP BY e.id
ORDER BY e.start_date DESC;

-- name: UpdateUserExperience :one
UPDATE user_experiences
SET
    company_name = COALESCE(sqlc.narg('company_name'), company_name),
    company_url = COALESCE(sqlc.narg('company_url'), company_url),
    role_title = COALESCE(sqlc.narg('role_title'), role_title),
    exp_location = COALESCE(sqlc.narg('exp_location'), exp_location),
    industry = COALESCE(sqlc.narg('industry'), industry),
    employment_type = COALESCE(sqlc.narg('employment_type'), employment_type),
    description = COALESCE(sqlc.narg('description'), description),
    achievements = COALESCE(sqlc.narg('achievements'), achievements),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = CASE
        WHEN sqlc.narg('is_current') = TRUE THEN NULL
        ELSE COALESCE(sqlc.narg('end_date'), end_date)
    END,
    is_current = COALESCE(sqlc.narg('is_current'), is_current),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteUserExperience :exec
DELETE FROM user_experiences
WHERE id = $1 AND user_id = $2;

-- name: AddSkillToExperience :one
INSERT INTO experience_skills (
    experience_id, skill_id
) VALUES ($1, $2)
RETURNING *;

-- name: GetOrCreateSkill :one
INSERT INTO skills (
    name
) VALUES ($1)
ON CONFLICT (name)
-- This is a "fake" update just to get the id back
DO UPDATE SET name = excluded.name
RETURNING id;

-- name: DeleteSkillsByExperienceID :exec
DELETE FROM experience_skills
WHERE experience_id = $1;
