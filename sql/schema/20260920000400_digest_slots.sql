-- +goose Up
ALTER TABLE email_digests ADD COLUMN digest_slot INTEGER NOT NULL DEFAULT 0 CHECK(digest_slot BETWEEN 0 AND 23);
ALTER TABLE email_digests DROP CONSTRAINT email_digests_user_id_digest_date_key;
ALTER TABLE email_digests ADD UNIQUE(user_id,digest_date,digest_slot);
-- +goose Down
-- Preserve delivery history: refuse rollback if multiple slots already exist on a day.
ALTER TABLE email_digests DROP CONSTRAINT email_digests_user_id_digest_date_digest_slot_key;
ALTER TABLE email_digests ADD UNIQUE(user_id,digest_date);
ALTER TABLE email_digests DROP COLUMN digest_slot;
