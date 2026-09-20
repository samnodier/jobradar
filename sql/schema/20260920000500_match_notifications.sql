-- +goose Up
CREATE TABLE match_notifications (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 read_at TIMESTAMPTZ,
 UNIQUE(user_id,job_id)
);
CREATE INDEX match_notifications_user_time_idx ON match_notifications(user_id,created_at DESC);
SELECT schedule_user_matches(id) FROM users;
-- +goose Down
DROP TABLE match_notifications;
