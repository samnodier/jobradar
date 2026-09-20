-- +goose Up
ALTER TABLE jobs ADD COLUMN last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX jobs_created_at_id_idx ON jobs (created_at DESC, id DESC);
CREATE INDEX matches_user_score_idx ON user_job_matches (user_id, match_score DESC);

-- PostgreSQL is the durable source of pending match work. Redis is its delivery mechanism.
CREATE TABLE pending_matches (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    revision BIGINT NOT NULL DEFAULT 1,
    enqueued_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, job_id)
);

-- +goose StatementBegin
CREATE FUNCTION schedule_user_matches(target_user UUID) RETURNS void LANGUAGE sql AS $$
    INSERT INTO pending_matches (user_id, job_id)
    SELECT target_user, id FROM jobs
    WHERE (created_by_user_id IS NULL OR created_by_user_id = target_user)
      AND EXISTS (SELECT 1 FROM users WHERE id = target_user)
    ON CONFLICT (user_id, job_id) DO UPDATE
    SET revision = pending_matches.revision + 1, enqueued_at = NULL;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION schedule_job_matches() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.updated_at IS NOT DISTINCT FROM OLD.updated_at THEN
        RETURN NEW;
    END IF;
    INSERT INTO pending_matches (user_id, job_id)
    SELECT id, NEW.id FROM users
    WHERE NEW.created_by_user_id IS NULL OR id = NEW.created_by_user_id
    ON CONFLICT (user_id, job_id) DO UPDATE
    SET revision = pending_matches.revision + 1, enqueued_at = NULL;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER jobs_schedule_matches AFTER INSERT OR UPDATE ON jobs
FOR EACH ROW EXECUTE FUNCTION schedule_job_matches();

-- +goose StatementBegin
CREATE FUNCTION schedule_profile_matches() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM schedule_user_matches(OLD.user_id);
        RETURN OLD;
    END IF;
    PERFORM schedule_user_matches(NEW.user_id);
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER skills_schedule_matches AFTER INSERT OR UPDATE OR DELETE ON user_skills
FOR EACH ROW EXECUTE FUNCTION schedule_profile_matches();
CREATE TRIGGER roles_schedule_matches AFTER INSERT OR UPDATE OR DELETE ON user_desired_roles
FOR EACH ROW EXECUTE FUNCTION schedule_profile_matches();
CREATE TRIGGER experiences_schedule_matches AFTER INSERT OR UPDATE OR DELETE ON user_experiences
FOR EACH ROW EXECUTE FUNCTION schedule_profile_matches();

-- +goose StatementBegin
CREATE FUNCTION schedule_experience_skill_matches() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE owner_id UUID;
BEGIN
    IF TG_OP = 'DELETE' THEN
        SELECT user_id INTO owner_id FROM user_experiences WHERE id = OLD.experience_id;
    ELSE
        SELECT user_id INTO owner_id FROM user_experiences WHERE id = NEW.experience_id;
    END IF;
    IF owner_id IS NOT NULL THEN PERFORM schedule_user_matches(owner_id); END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER experience_skills_schedule_matches AFTER INSERT OR DELETE ON experience_skills
FOR EACH ROW EXECUTE FUNCTION schedule_experience_skill_matches();

CREATE TABLE email_digests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    digest_date DATE NOT NULL,
    recipient TEXT NOT NULL,
    sender TEXT NOT NULL,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    html TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sending','sent','uncertain','failed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    first_attempt_at TIMESTAMPTZ,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    provider_message_id TEXT,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ,
    UNIQUE (user_id, digest_date)
);
CREATE TABLE email_digest_jobs (
    digest_id UUID NOT NULL REFERENCES email_digests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, job_id)
);

-- Backfill once: migration preserves existing jobs and makes old missed work recoverable.
INSERT INTO pending_matches (user_id, job_id)
SELECT u.id, j.id FROM users u JOIN jobs j
ON j.created_by_user_id IS NULL OR j.created_by_user_id = u.id
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE email_digest_jobs;
DROP TABLE email_digests;
DROP TRIGGER experience_skills_schedule_matches ON experience_skills;
DROP FUNCTION schedule_experience_skill_matches();
DROP TRIGGER experiences_schedule_matches ON user_experiences;
DROP TRIGGER roles_schedule_matches ON user_desired_roles;
DROP TRIGGER skills_schedule_matches ON user_skills;
DROP FUNCTION schedule_profile_matches();
DROP TRIGGER jobs_schedule_matches ON jobs;
DROP FUNCTION schedule_job_matches();
DROP FUNCTION schedule_user_matches(UUID);
DROP TABLE pending_matches;
DROP INDEX matches_user_score_idx;
DROP INDEX jobs_created_at_id_idx;
ALTER TABLE jobs DROP COLUMN last_seen_at;
