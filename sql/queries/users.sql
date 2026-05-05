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
    is_admin,
    created_at,
    updated_at
FROM users
WHERE id = $1
ORDER BY id
LIMIT 1;

-- name: GetUserByProviderIdentity :one
SELECT
    u.id,
    u.email,
    u.username,
    u.full_name,
    u.avatar_url,
    u.is_admin,
    u.created_at,
    u.updated_at
FROM users AS u
INNER JOIN user_accounts AS ua ON u.id = ua.user_id
WHERE ua.auth_provider = $1 AND ua.auth_provider_id = $2
ORDER BY u.id
LIMIT 1;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;
