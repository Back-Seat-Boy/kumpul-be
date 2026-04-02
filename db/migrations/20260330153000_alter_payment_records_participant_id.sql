-- +migrate Up

ALTER TABLE payment_records ADD COLUMN participant_id UUID;

UPDATE payment_records pr
SET participant_id = ptn.id
FROM payments p, participants ptn
WHERE pr.payment_id = p.id
  AND ptn.event_id = p.event_id
  AND ptn.user_id = pr.user_id;

ALTER TABLE payment_records
    ALTER COLUMN participant_id SET NOT NULL;

ALTER TABLE payment_records
    ADD CONSTRAINT payment_records_participant_id_fkey
    FOREIGN KEY (participant_id) REFERENCES participants(id) ON DELETE CASCADE;

CREATE INDEX idx_payment_records_participant_id ON payment_records(participant_id);

DROP INDEX IF EXISTS idx_payment_records_user_id;

ALTER TABLE payment_records DROP COLUMN user_id;


-- +migrate Down

ALTER TABLE payment_records ADD COLUMN user_id UUID;

UPDATE payment_records pr
SET user_id = ptn.user_id
FROM participants ptn
WHERE pr.participant_id = ptn.id;

DELETE FROM payment_records pr
USING participants ptn
WHERE pr.participant_id = ptn.id
  AND ptn.user_id IS NULL;

CREATE INDEX idx_payment_records_user_id ON payment_records(user_id);

ALTER TABLE payment_records
    DROP CONSTRAINT IF EXISTS payment_records_participant_id_fkey;

DROP INDEX IF EXISTS idx_payment_records_participant_id;

ALTER TABLE payment_records DROP COLUMN participant_id;

ALTER TABLE payment_records
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE payment_records
    ADD CONSTRAINT payment_records_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id);
