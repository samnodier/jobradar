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
            JSON_BUILD_OBJECT('id', s.id, 'skill_name', s.skill_name)
        ) FILTER (WHERE s.id IS NOT NULL), '[]'
    ) AS skills
FROM user_experiences AS e
LEFT JOIN experience_skills AS es ON e.id = es.experience_id
LEFT JOIN skills AS s ON es.skill_id = s.id
WHERE e.user_id = $1
GROUP BY e.id
ORDER BY e.start_date DESC;

-- name: GetExperienceByID :one
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
            JSON_BUILD_OBJECT('id', s.id, 'skill_name', s.skill_name)
        ) FILTER (WHERE s.id IS NOT NULL), '[]'
    ) AS skills
FROM user_experiences AS e
LEFT JOIN experience_skills AS es ON e.id = es.experience_id
LEFT JOIN skills AS s ON es.skill_id = s.id
WHERE e.id = $1 AND e.user_id = $2
GROUP BY e.id;


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
        WHEN sqlc.narg('is_current')::boolean = TRUE THEN NULL
        ELSE COALESCE(sqlc.narg('end_date')::date, end_date)
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
    skill_name
) VALUES ($1)
ON CONFLICT (skill_name)
-- This is a "fake" update just to get the id back
DO UPDATE SET skill_name = excluded.skill_name
RETURNING id;

-- name: DeleteSkillsByExperienceID :exec
DELETE FROM experience_skills
WHERE experience_id = $1;

-- name: CreateUserEducation :one
INSERT INTO user_education (
    user_id,
    institution_name,
    degree_type,
    degree_name,
    field_of_study,
    start_date,
    end_date,
    is_current,
    description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetEducationsByUserID :many
SELECT
    id,
    user_id,
    institution_name,
    degree_type,
    degree_name,
    field_of_study,
    start_date,
    end_date,
    is_current,
    description,
    created_at,
    updated_at
FROM user_education
WHERE user_id = $1
ORDER BY start_date DESC;

-- name: UpdateUserEducation :one
UPDATE user_education
SET
    institution_name
    = COALESCE(sqlc.narg('institution_name'), institution_name),
    degree_type = COALESCE(sqlc.narg('degree_type'), degree_type),
    degree_name = COALESCE(sqlc.narg('degree_name'), degree_name),
    field_of_study = COALESCE(sqlc.narg('field_of_study'), field_of_study),
    start_date = COALESCE(sqlc.narg('start_date')::date, start_date),
    end_date = CASE
        WHEN sqlc.narg('is_current')::boolean = TRUE THEN NULL
        ELSE COALESCE(sqlc.narg('end_date')::date, end_date)
    END,
    is_current = COALESCE(sqlc.narg('is_current')::boolean, is_current),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteUserEducation :exec
DELETE FROM user_education
WHERE id = $1 AND user_id = $2;

-- name: CreateUserProject :one
INSERT INTO user_projects (
    user_id,
    title,
    role_title,
    description,
    impact,
    project_url,
    repository_url,
    start_date,
    end_date,
    is_current,
    is_featured
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetProjectsByUserID :many
SELECT
    id,
    user_id,
    title,
    role_title,
    description,
    impact,
    project_url,
    repository_url,
    start_date,
    end_date,
    is_current,
    is_featured,
    created_at,
    updated_at
FROM user_projects
WHERE user_id = $1
ORDER BY start_date DESC;

-- name: UpdateUserProject :one
UPDATE user_projects
SET
    title = COALESCE(sqlc.narg('title'), title),
    role_title = COALESCE(sqlc.narg('role_title'), role_title),
    description = COALESCE(sqlc.narg('description'), description),
    impact = COALESCE(sqlc.narg('impact'), impact),
    project_url = COALESCE(sqlc.narg('project_url'), project_url),
    repository_url = COALESCE(sqlc.narg('repository_url'), repository_url),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = CASE
        WHEN sqlc.narg('is_current')::boolean = TRUE THEN NULL
        ELSE COALESCE(sqlc.narg('end_date')::date, end_date)
    END,
    is_current = COALESCE(sqlc.narg('is_current'), is_current),
    is_featured = COALESCE(sqlc.narg('is_featured'), is_featured),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteUserProject :exec
DELETE FROM user_projects
WHERE id = $1 AND user_id = $2;

-- name: CreateUserCertification :one
INSERT INTO user_certifications (
    user_id,
    certification_name,
    issuing_organization,
    issue_date,
    expiration_date,
    does_not_expire,
    credential_id,
    credential_url,
    is_in_progress
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetCertificationsByUserID :many
SELECT
    id,
    user_id,
    certification_name,
    issuing_organization,
    issue_date,
    expiration_date,
    does_not_expire,
    credential_id,
    credential_url,
    is_in_progress,
    created_at,
    updated_at
FROM user_certifications
WHERE user_id = $1
ORDER BY issue_date DESC;

-- name: UpdateUserCertification :one
UPDATE user_certifications
SET
    certification_name
    = COALESCE(sqlc.narg('certification_name'), certification_name),
    issuing_organization
    = COALESCE(sqlc.narg('issuing_organization'), issuing_organization),
    issue_date = COALESCE(sqlc.narg('issue_date'), issue_date),
    expiration_date = COALESCE(sqlc.narg('expiration_date'), expiration_date),
    does_not_expire = COALESCE(sqlc.narg('does_not_expire'), does_not_expire),
    credential_id = COALESCE(sqlc.narg('credential_id'), credential_id),
    credential_url = COALESCE(sqlc.narg('credential_url'), credential_url),
    is_in_progress = COALESCE(sqlc.narg('is_in_progress'), is_in_progress),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteUserCertification :exec
DELETE FROM user_certifications
WHERE id = $1 AND user_id = $2;

-- name: CreateUserLanguage :one
INSERT INTO user_languages (
    user_id,
    user_language,
    proficiency
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetLanguagesByUserID :many
SELECT
    user_id,
    user_language,
    proficiency,
    created_at,
    updated_at
FROM user_languages
WHERE user_id = $1
ORDER BY proficiency DESC;

-- name: UpdateUserLanguage :one
UPDATE user_languages
SET
    proficiency = COALESCE(sqlc.narg('proficiency'), proficiency),
    updated_at = NOW()
WHERE user_id = $1 AND user_language = $2
RETURNING *;

-- name: DeleteUserLanguage :exec
DELETE FROM user_languages
WHERE user_id = $1 AND user_language = $2;

-- name: GetUserSkillsByUserID :many
SELECT
    s.id AS skill_id,
    s.skill_name,
    s.category AS skill_category,
    us.proficiency,
    us.years_experience,
    us.is_featured,
    us.endorsed_by_ai,
    us.display_order
FROM user_skills AS us
INNER JOIN skills AS s ON us.skill_id = s.id
WHERE us.user_id = $1
ORDER BY us.display_order ASC, s.skill_name ASC;

-- name: AddSkillToUser :one
INSERT INTO user_skills (
    user_id,
    skill_id,
    proficiency,
    years_experience,
    is_featured,
    endorsed_by_ai,
    display_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ClearUserSkills :exec
DELETE FROM user_skills
WHERE user_id = $1;

-- name: GetUserDesiredLocations :many
SELECT
    location_name,
    is_remote,
    priority
FROM user_desired_locations
WHERE user_id = $1
ORDER BY priority ASC, location_name ASC;

-- name: AddDesiredLocation :one
INSERT INTO user_desired_locations (
    user_id,
    location_name,
    is_remote,
    priority
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ClearUserDesiredLocations :exec
DELETE FROM user_desired_locations
WHERE user_id = $1;

-- name: GetUserDesiredRoles :many                              
SELECT
    role_title,
    priority
FROM user_desired_roles
WHERE user_id = $1
ORDER BY priority ASC, role_title ASC;

-- name: AddDesiredRole :one
INSERT INTO user_desired_roles (
    user_id,
    role_title,
    priority
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: ClearUserDesiredRoles :exec
DELETE FROM user_desired_roles
WHERE user_id = $1;
