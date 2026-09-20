-- +goose Up
CREATE TABLE user_match_settings (
 user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 min_score INTEGER NOT NULL DEFAULT 60 CHECK (min_score BETWEEN 0 AND 100),
 work_mode TEXT NOT NULL DEFAULT 'any' CHECK (work_mode IN ('any','remote','onsite')),
 include_unknown BOOLEAN NOT NULL DEFAULT true,
 search_country TEXT NOT NULL DEFAULT 'us',
 email_enabled BOOLEAN NOT NULL DEFAULT true,
 notifications_enabled BOOLEAN NOT NULL DEFAULT true,
 digest_timezone TEXT NOT NULL DEFAULT 'America/New_York',
 digest_hours INTEGER[] NOT NULL DEFAULT '{8,18}' CHECK (cardinality(digest_hours) BETWEEN 1 AND 2 AND 0 <= ALL(digest_hours) AND 23 >= ALL(digest_hours)),
 digest_max_jobs INTEGER NOT NULL DEFAULT 10 CHECK (digest_max_jobs BETWEEN 1 AND 25)
);
CREATE TRIGGER match_settings_schedule_matches AFTER INSERT OR UPDATE OR DELETE ON user_match_settings
FOR EACH ROW EXECUTE FUNCTION schedule_profile_matches();
CREATE TRIGGER locations_schedule_matches AFTER INSERT OR UPDATE OR DELETE ON user_desired_locations
FOR EACH ROW EXECUTE FUNCTION schedule_profile_matches();
-- +goose StatementBegin
CREATE FUNCTION schedule_user_preference_matches() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF (NEW.min_salary,NEW.max_salary,NEW.salary_currency,NEW.preferred_job_types,NEW.notify_jobs)
 IS DISTINCT FROM (OLD.min_salary,OLD.max_salary,OLD.salary_currency,OLD.preferred_job_types,OLD.notify_jobs) THEN
  PERFORM schedule_user_matches(NEW.id);
 END IF;
 RETURN NEW;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER user_preferences_schedule_matches AFTER UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION schedule_user_preference_matches();
-- +goose Down
DROP TRIGGER user_preferences_schedule_matches ON users;
DROP FUNCTION schedule_user_preference_matches();
DROP TRIGGER locations_schedule_matches ON user_desired_locations;
DROP TABLE user_match_settings;
