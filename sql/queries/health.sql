-- name: UpdateServiceHealth :one
INSERT INTO service_health (
    service_name,
    last_run_at,
    last_success_at,
    service_status,
    last_error,
    job_count_last_run
) VALUES (
    $1, NOW(), $2, $3, $4, $5
)
ON CONFLICT (service_name) DO UPDATE
    SET
        last_run_at = excluded.last_run_at,
        last_success_at = COALESCE(
            excluded.last_success_at,       -- new value (maybe be null)
            service_health.last_success_at  -- existing value (fallback)
        ),
        service_status = excluded.service_status,
        last_error = excluded.last_error,
        job_count_last_run = excluded.job_count_last_run,
        updated_at = NOW()
RETURNING *;
