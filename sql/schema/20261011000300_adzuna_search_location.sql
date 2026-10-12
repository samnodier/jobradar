-- +goose Up
ALTER TABLE user_match_settings ADD COLUMN search_location TEXT NOT NULL DEFAULT '' CHECK (length(search_location)<=100);
-- Search market location is distinct from geographic matching preferences.
-- +goose Down
ALTER TABLE user_match_settings DROP COLUMN search_location;
