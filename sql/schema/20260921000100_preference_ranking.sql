-- +goose Up
ALTER TABLE jobs ADD COLUMN industry TEXT, ADD COLUMN category TEXT, ADD COLUMN company_stage TEXT;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION schedule_user_preference_matches() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF (NEW.min_salary,NEW.max_salary,NEW.salary_currency,NEW.preferred_job_types,NEW.notify_jobs,NEW.preferred_industries,NEW.company_stage_preference)
 IS DISTINCT FROM (OLD.min_salary,OLD.max_salary,OLD.salary_currency,OLD.preferred_job_types,OLD.notify_jobs,OLD.preferred_industries,OLD.company_stage_preference) THEN
  PERFORM schedule_user_matches(NEW.id);
 END IF;
 RETURN NEW;
END;
$$;
-- +goose StatementEnd
SELECT schedule_user_matches(id) FROM users;
-- +goose Down
ALTER TABLE jobs DROP COLUMN industry, DROP COLUMN category, DROP COLUMN company_stage;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION schedule_user_preference_matches() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF (NEW.min_salary,NEW.max_salary,NEW.salary_currency,NEW.preferred_job_types,NEW.notify_jobs)
 IS DISTINCT FROM (OLD.min_salary,OLD.max_salary,OLD.salary_currency,OLD.preferred_job_types,OLD.notify_jobs) THEN
  PERFORM schedule_user_matches(NEW.id);
 END IF;
 RETURN NEW;
END;
$$;
-- +goose StatementEnd
