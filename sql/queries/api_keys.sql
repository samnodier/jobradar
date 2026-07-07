-- name: UpsertUserAPIKey :exec
INSERT INTO user_api_keys (user_id, provider, encrypted_key)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, provider)
DO UPDATE SET encrypted_key = EXCLUDED.encrypted_key, updated_at = NOW();

-- name: GetUserAPIKey :one
SELECT encrypted_key
FROM user_api_keys
WHERE user_id = $1 AND provider = $2;

-- name: ListUserAPIKeyProviders :many
SELECT provider
FROM user_api_keys
WHERE user_id = $1
ORDER BY provider;

-- name: DeleteUserAPIKey :exec
DELETE FROM user_api_keys
WHERE user_id = $1 AND provider = $2;
