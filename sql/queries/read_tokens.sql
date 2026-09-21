-- name: CreateReadToken :one
INSERT INTO read_api_tokens(user_id,name,token_hash,expires_at) VALUES ($1,$2,$3,$4)
RETURNING id,name,created_at,expires_at;
-- name: ListReadTokens :many
SELECT id,name,created_at,expires_at,revoked_at FROM read_api_tokens WHERE user_id=$1 ORDER BY created_at DESC;
-- name: RevokeReadToken :execrows
UPDATE read_api_tokens SET revoked_at=now() WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL;
-- name: AuthenticateReadToken :one
SELECT u.id,u.email FROM read_api_tokens t JOIN users u ON u.id=t.user_id
WHERE t.token_hash=$1 AND t.revoked_at IS NULL AND t.expires_at>now();
