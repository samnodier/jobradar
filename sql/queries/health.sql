-- name: UpdateServiceHealth :one
INSERT INTO service_health (
    service_name, last_run_at, last_success_at, status, last_error, job_count_last_run
) VALUES (
    $1, NOW(), $2, $3, $4, $5
)
ON CONFLICT (service_name) DO UPDATE SET
   last_run_at = EXCLUDED.last_run_at,
   last_success_at = COALESCE(
        EXCLUDED.last_success_at,       -- new value (maybe be null)
        service_health.last_success_at  -- existing value (fallback)
    ),
   status = EXCLUDED.status,
   last_error = EXCLUDED.last_error,
   job_count_last_run = EXCLUDED.job_count_last_run,
   updated_at = NOW()
RETURNING *;
