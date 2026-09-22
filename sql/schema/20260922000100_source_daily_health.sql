-- +goose Up
CREATE TABLE source_health_daily (
    service_name TEXT NOT NULL REFERENCES service_health(service_name) ON DELETE CASCADE,
    day DATE NOT NULL,
    successes INTEGER NOT NULL DEFAULT 0 CHECK (successes >= 0),
    failures INTEGER NOT NULL DEFAULT 0 CHECK (failures >= 0),
    PRIMARY KEY (service_name, day)
);
CREATE INDEX source_health_daily_day_idx ON source_health_daily(day);

-- Record only actual fetch observations; do not fabricate historical uptime.
-- +goose StatementBegin
CREATE FUNCTION record_source_health_day() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO source_health_daily(service_name, day, successes, failures)
    VALUES (NEW.service_name, (NEW.last_run_at AT TIME ZONE 'UTC')::date,
            CASE WHEN NEW.service_status = 'healthy' THEN 1 ELSE 0 END,
            CASE WHEN NEW.service_status = 'healthy' THEN 0 ELSE 1 END)
    ON CONFLICT (service_name, day) DO UPDATE SET
        successes = source_health_daily.successes + EXCLUDED.successes,
        failures = source_health_daily.failures + EXCLUDED.failures;
    DELETE FROM source_health_daily WHERE day < (NOW() AT TIME ZONE 'UTC')::date - 29;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER source_health_daily_observation
AFTER INSERT OR UPDATE OF last_run_at ON service_health
FOR EACH ROW EXECUTE FUNCTION record_source_health_day();

-- +goose Down
DROP TRIGGER source_health_daily_observation ON service_health;
DROP FUNCTION record_source_health_day();
DROP TABLE source_health_daily;
