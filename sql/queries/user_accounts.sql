-- name: CreateUserAccount :one
INSERT INTO user_accounts (
    user_id, auth_provider, auth_provider_id, access_token
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;
