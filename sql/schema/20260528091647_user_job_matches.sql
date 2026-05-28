-- +goose Up
CREATE TABLE IF NOT EXISTS user_job_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    match_score INTEGER,
    title_score FLOAT,
    skill_score FLOAT,
    experience_score FLOAT,
    matched_skills TEXT [] DEFAULT '{}',
    missing_skills TEXT [] DEFAULT '{}',
    ai_summary TEXT,
    is_enriched BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, job_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_job_matches;
