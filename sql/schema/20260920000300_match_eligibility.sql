-- +goose Up
ALTER TABLE user_job_matches ADD COLUMN preference_eligible BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE user_job_matches ADD COLUMN preference_notes TEXT[] NOT NULL DEFAULT '{}';
SELECT schedule_user_matches(id) FROM users;
-- +goose Down
ALTER TABLE user_job_matches DROP COLUMN preference_notes, DROP COLUMN preference_eligible;
