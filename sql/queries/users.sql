-- name: CreateUser :one
INSERT INTO users (
    email, username, full_name, avatar_url
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByID :one
SELECT
    id,
    email,
    username,
    full_name,
    avatar_url,
    phone,
    user_location,
    website_url,
    linkedin_url,
    github_url,
    headline,
    user_summary,
    availability,
    min_salary,
    max_salary,
    salary_currency,
    years_of_experience,
    is_admin,
    created_at,
    updated_at
FROM users
WHERE id = $1;

-- name: GetUserByProviderIdentity :one
SELECT
    u.id,
    u.email,
    u.username,
    u.full_name,
    u.avatar_url,
    u.phone,
    u.user_location,
    u.website_url,
    u.linkedin_url,
    u.github_url,
    u.headline,
    u.user_summary,
    u.availability,
    u.min_salary,
    u.max_salary,
    u.salary_currency,
    u.years_of_experience,
    u.is_admin,
    u.created_at,
    u.updated_at
FROM users AS u
INNER JOIN user_accounts AS ua ON u.id = ua.user_id
WHERE ua.auth_provider = $1 AND ua.auth_provider_id = $2;

-- name: UpdateUser :one
UPDATE users
SET
    username = COALESCE(sqlc.arg('username'), username),
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    phone = COALESCE(sqlc.narg('phone'), phone),
    user_location = COALESCE(sqlc.narg('user_location'), user_location),
    website_url = COALESCE(sqlc.narg('website_url'), website_url),
    linkedin_url = COALESCE(sqlc.narg('linkedin_url'), linkedin_url),
    github_url = COALESCE(sqlc.narg('github_url'), github_url),
    -- Identity
    headline = COALESCE(sqlc.narg('headline'), headline),
    user_summary = COALESCE(sqlc.narg('user_summary'), user_summary),
    -- Career Preferences
    availability = COALESCE(sqlc.narg('availability'), availability),
    min_salary = COALESCE(sqlc.narg('min_salary'), min_salary),
    max_salary = COALESCE(sqlc.narg('max_salary'), max_salary),
    salary_currency = COALESCE(sqlc.narg('salary_currency'), salary_currency),
    years_of_experience
    = COALESCE(sqlc.narg('years_of_experience'), years_of_experience),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;
