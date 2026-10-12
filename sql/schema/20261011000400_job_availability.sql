-- +goose Up
CREATE TABLE job_availability (
 job_id UUID PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE,
 status TEXT NOT NULL CHECK (status IN ('reachable','unavailable','unknown')),
 response_code INTEGER NOT NULL DEFAULT 0,
 checked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE job_availability;
