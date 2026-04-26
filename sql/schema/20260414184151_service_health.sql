-- +goose Up
CREATE TABLE IF NOT EXISTS service_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT UNIQUE NOT NULL,  -- eg. 'remoteok', 'adzuna'
    last_run_at TIMESTAMPTZ NOT NULL,
    last_success_at TIMESTAMPTZ,
    status TEXT NOT NULL,               -- 'healthy', 'degraded', 'failing'
    last_error TEXT,                    -- Store the error message
    job_count_last_run INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS service_health;
