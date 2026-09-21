-- name: GetPasswordLogin :one
SELECT u.id,u.email,u.auth_version,p.password_hash FROM users u
JOIN user_passwords p ON p.user_id=u.id WHERE u.email=$1;
-- name: GetAuthIdentity :one
SELECT id,email,auth_version FROM users WHERE id=$1;
-- name: GetAuthIdentityByEmail :one
SELECT id,email,auth_version FROM users WHERE email=$1;
-- name: LockAuthIdentity :one
SELECT id,email,auth_version FROM users WHERE id=$1 FOR UPDATE;
-- name: GetPasswordHash :one
SELECT password_hash FROM user_passwords WHERE user_id=$1;
-- name: SetPasswordHash :exec
INSERT INTO user_passwords(user_id,password_hash)VALUES($1,$2)
ON CONFLICT(user_id) DO UPDATE SET password_hash=excluded.password_hash,updated_at=now();
-- name: InvalidateUserAuthentication :one
UPDATE users SET auth_version=auth_version+1 WHERE id=$1 RETURNING auth_version;
-- name: RevokeUserReadTokens :exec
UPDATE read_api_tokens SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL;
-- name: DeletePasswordSetupTokens :exec
DELETE FROM password_setup_tokens WHERE user_id=$1;
-- name: CreatePasswordSetupToken :exec
INSERT INTO password_setup_tokens(token_hash,user_id,expires_at)VALUES($1,$2,now()+interval '15 minutes');
-- name: GetPasswordSetupToken :one
SELECT user_id FROM password_setup_tokens WHERE token_hash=$1 AND expires_at>now();
-- name: ConsumePasswordSetupToken :one
DELETE FROM password_setup_tokens WHERE token_hash=$1 AND expires_at>now() RETURNING user_id;
