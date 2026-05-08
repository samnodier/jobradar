-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    username TEXT UNIQUE,
    full_name TEXT,
    avatar_url TEXT,
    -- Contact info
    phone TEXT,
    user_location TEXT,
    website_url TEXT,
    linkedin_url TEXT,
    github_url TEXT,
    -- Identity
    headline TEXT,
    user_summary TEXT,
    -- Career Preferences
    availability TEXT DEFAULT 'immediate',
    min_salary INTEGER,
    max_salary INTEGER,
    salary_currency TEXT DEFAULT 'USD',
    years_of_experience INTEGER,
    -- Meta
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS citext;
