-- +goose Up
-- Add the ownership column
ALTER TABLE jobs
ADD COLUMN created_by_user_id uuid
REFERENCES users (id)
ON DELETE CASCADE;

-- Drop the old unique constraint
ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_external_id_job_source_key;

-- Drop the global source_url unique constraint
ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_source_url_key;

-- Create two new partial unique indexes
CREATE UNIQUE INDEX jobs_external_id_job_source
ON jobs (external_id, job_source)
WHERE created_by_user_id IS NULL;

CREATE UNIQUE INDEX jobs_created_by_user_id_source_url
ON jobs (created_by_user_id, source_url)
WHERE created_by_user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS jobs_created_by_user_id_source_url;
DROP INDEX IF EXISTS jobs_external_id_job_source;

ALTER TABLE jobs
ADD CONSTRAINT jobs_source_url_key UNIQUE (source_url);

ALTER TABLE jobs
ADD CONSTRAINT jobs_external_id_job_source_key UNIQUE (external_id, job_source);

ALTER TABLE jobs
DROP COLUMN created_by_user_id;
