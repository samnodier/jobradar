-- +goose Up
CREATE TABLE IF NOT EXISTS user_desired_roles (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_title TEXT NOT NULL,
    priority INTEGER DEFAULT 0,
    PRIMARY KEY (user_id, role_title)
);

-- +goose Down
DROP TABLE IF EXISTS user_desired_roles;
