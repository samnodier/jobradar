-- name: CreateUserAccount :one
INSERT INTO user_accounts (
    user_id, auth_provider, auth_provider_id, access_token
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: HasGitHubAccount :one
SELECT EXISTS(SELECT 1 FROM user_accounts WHERE user_id=$1 AND auth_provider='github' AND revoked_at IS NULL);
