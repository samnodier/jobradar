-- +goose Up
CREATE TABLE user_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (provider IN ('gemini', 'groq')),
    encrypted_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, provider)
);

-- Carry existing Gemini ciphertexts into the new table
-- before dropping the column
INSERT INTO user_api_keys (user_id, provider, encrypted_key)
SELECT
    id,
    'gemini',
    encrypted_gemini_api_key
FROM users
WHERE encrypted_gemini_api_key IS NOT NULL;

ALTER TABLE users DROP COLUMN encrypted_gemini_api_key;

-- +goose Down
ALTER TABLE users ADD COLUMN encrypted_gemini_api_key TEXT;

UPDATE users
SET encrypted_gemini_api_key = k.encrypted_key
FROM user_api_keys AS k
WHERE k.user_id = users.id AND k.provider = 'gemini';

DROP TABLE user_api_keys;
