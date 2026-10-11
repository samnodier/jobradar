-- +goose Up
-- Experience scoring now uses the best single experience (BestExperienceMatch). Stored
-- scores were computed with the diluted union formula, so queue every user's jobs for
-- re-scoring; the durable pending-match dispatcher recomputes them with the new code.
SELECT schedule_user_matches(id) FROM users;

-- +goose Down
-- Nothing to undo: scores are always recomputed from current data and code.
SELECT 1;
