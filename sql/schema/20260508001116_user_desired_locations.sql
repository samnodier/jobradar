-- +goose Up
CREATE TABLE IF NOT EXISTS user_desired_locations (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    location_name TEXT NOT NULL,
    is_remote BOOLEAN DEFAULT FALSE,
    priority INTEGER DEFAULT 0,
    PRIMARY KEY (user_id, location_name)
);

-- +goose Down
DROP TABLE IF EXISTS user_desired_locations;
