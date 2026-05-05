-- +goose Up
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    application_status TEXT NOT NULL CHECK (
        application_status IN (
            'applied',
            'recruiter_screen',
            'interview',
            'offer',
            'accepted',
            'rejected',
            'withdrawn'
        )
    ),
    applied_at TIMESTAMPTZ NULL,
    last_status_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    follow_up_at TIMESTAMPTZ NULL,
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, job_id)
);

-- +goose Down
DROP TABLE IF EXISTS applications;
