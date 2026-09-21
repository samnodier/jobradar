-- +goose Up
ALTER TABLE users ADD COLUMN auth_version BIGINT NOT NULL DEFAULT 0;
CREATE TABLE user_passwords (
 user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 password_hash TEXT NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE password_setup_tokens (
 token_hash TEXT PRIMARY KEY,
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL
);
-- +goose Down
DROP TABLE password_setup_tokens;
DROP TABLE user_passwords;
ALTER TABLE users DROP COLUMN auth_version;
