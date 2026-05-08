-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id TEXT NOT NULL,
    job_source TEXT NOT NULL,
    title TEXT NOT NULL,
    company_name TEXT NOT NULL,
    description TEXT,
    source_url TEXT UNIQUE NOT NULL,
    salary_min INTEGER,
    salary_max INTEGER,
    currency VARCHAR(10),
    job_location TEXT,
    is_remote BOOLEAN,
    job_status TEXT DEFAULT 'new',
    employment_type TEXT,
    experience_level TEXT,
    skills TEXT [],
    posted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    logo_url TEXT,
    -- AI matching fields
    match_score INTEGER,
    ai_summary TEXT,
    matched_skills TEXT [],
    missing_skills TEXT [],
    UNIQUE (external_id, job_source)
);

-- +goose Down
DROP TABLE IF EXISTS jobs;
