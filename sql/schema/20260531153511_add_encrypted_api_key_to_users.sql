-- +goose Up
ALTER TABLE users ADD COLUMN encrypted_gemini_api_key TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN encrypted_gemini_api_key;
