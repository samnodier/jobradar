-- name: CreateUser :one
INSERT INTO users (
    email, username, name, avatar_url
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByProviderIdentity :one
SELECT u.* FROM users u
JOIN user_accounts ua ON ua.user_id = u.id
WHERE ua.provider = $1 AND ua.provider_id = $2
LIMIT 1;
