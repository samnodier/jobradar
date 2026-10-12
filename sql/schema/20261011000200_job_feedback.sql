-- +goose Up
CREATE TABLE job_feedback (
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
 sentiment SMALLINT NOT NULL CHECK (sentiment IN (-1,1)),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY (user_id,job_id)
);
-- Use the existing durable outbox so feedback and rematching commit together.
CREATE TRIGGER feedback_rematch AFTER INSERT OR UPDATE OR DELETE ON job_feedback
 FOR EACH ROW EXECUTE FUNCTION schedule_profile_matches();

-- +goose Down
DROP TABLE job_feedback;
SELECT schedule_user_matches(id) FROM users;
