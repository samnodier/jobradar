-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS jobs (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       external_id TEXT NOT NULL,
       source TEXT NOT NULL,
       title TEXT NOT NULL,
       company TEXT NOT NULL,
       description TEXT,
       url TEXT UNIQUE NOT NULL,
       salary_min INTEGER,
       salary_max INTEGER,
       currency VARCHAR(10),
       location TEXT,
       is_remote BOOLEAN,
       status TEXT DEFAULT 'new',
       employment_type TEXT,
       experience_level TEXT,
       skills TEXT[],
       posted_at TIMESTAMPTZ,
       expires_at TIMESTAMPTZ,
       created_at TIMESTAMPTZ DEFAULT NOW(),
       updated_at TIMESTAMPTZ DEFAULT NOW(),
       logo_url TEXT,
       UNIQUE (external_id, source)
);

-- +goose Down
DROP TABLE IF EXISTS jobs;
