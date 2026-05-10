-- +goose Up
CREATE TABLE IF NOT EXISTS user_experiences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid not null references users (id) on delete cascade,
    company_name TEXT NOT NULL,
    company_url TEXT,
    role_title TEXT NOT NULL,
    exp_location TEXT,
    industry TEXT,
    employment_type TEXT, -- 'full-time', 'contract', 'freelance'
    description TEXT, -- The "what I did"
    achievements TEXT [], -- The "what I won"
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT check_dates CHECK (end_date IS NULL OR end_date >= start_date),
    CONSTRAINT check_current_exp CHECK (
        NOT (is_current = TRUE AND end_date IS NOT NULL)
    )
);

CREATE TABLE user_education (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    institution_name TEXT NOT NULL,
    degree_type TEXT, -- 'bachelor', 'master', 'bootcamp'
    degree_name TEXT, -- 'Computer Science'
    field_of_study TEXT,
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE,
    description TEXT,
    is_highlighted BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT check_current_edu CHECK (
        NOT (is_current = TRUE AND end_date IS NOT NULL)
    )
);

CREATE TABLE user_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    role_title TEXT,
    description TEXT,
    impact TEXT, -- Quantified results
    project_url TEXT,
    repository_url TEXT,
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE user_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    issuing_organization TEXT NOT NULL,
    issue_date DATE,
    expiration_date DATE,
    does_not_expire BOOLEAN DEFAULT FALSE,
    credential_id TEXT,
    credential_url TEXT,
    is_in_progress BOOLEAN DEFAULT FALSE,
    -- Important for region-specific licenses like RN or CPA
    location TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);


CREATE TABLE IF NOT EXISTS user_languages (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- 'native', 'fluent', 'professional'
    language TEXT NOT NULL, proficiency TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, language)
);


-- Skill Junctions for Resume Sections
CREATE TABLE IF NOT EXISTS experience_skills (
    experience_id UUID NOT NULL REFERENCES user_experiences (
        id
    ) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (experience_id, skill_id)
);

CREATE TABLE IF NOT EXISTS project_skills (
    project_id UUID NOT NULL REFERENCES user_projects (id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (project_id, skill_id)
);

-- +goose Down
DROP TABLE IF EXISTS project_skills;
DROP TABLE IF EXISTS experience_skills;
DROP TABLE IF EXISTS user_languages;
DROP TABLE IF EXISTS user_certifications;
DROP TABLE IF EXISTS user_projects;
DROP TABLE IF EXISTS user_education;
DROP TABLE IF EXISTS user_experiences;
