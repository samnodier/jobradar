-- +goose Up
CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_name CITEXT NOT NULL UNIQUE, -- 'Go', 'Vue.js'
    category TEXT, -- 'Language', 'Framework', 'Database'
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_skills (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills (id) ON DELETE CASCADE,
    proficiency TEXT, -- 'beginner', 'intermediate', 'advanced', 'expert'
    years_experience INTEGER DEFAULT 0,
    is_featured BOOLEAN DEFAULT FALSE,
    endorsed_by_ai BOOLEAN DEFAULT FALSE,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, skill_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_skills;
DROP TABLE IF EXISTS skills;
