-- +migrate Up

ALTER TABLE payments RENAME COLUMN split_amount TO base_split;

ALTER TABLE payment_records
    ADD COLUMN amount INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN note TEXT;

UPDATE payment_records
SET amount = paid_amount;


-- +migrate Down

UPDATE payment_records
SET paid_amount = amount;

ALTER TABLE payment_records
    DROP COLUMN IF EXISTS note,
    DROP COLUMN IF EXISTS amount;

ALTER TABLE payments RENAME COLUMN base_split TO split_amount;
