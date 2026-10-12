-- name: RecordJobAvailability :exec
INSERT INTO job_availability(job_id,status,response_code) VALUES($1,$2,$3)
ON CONFLICT(job_id) DO UPDATE SET status=excluded.status,response_code=excluded.response_code,checked_at=now();

-- name: GetJobAvailability :one
SELECT status,response_code,checked_at FROM job_availability WHERE job_id=$1;

