-- +migrate Up

ALTER TABLE payment_records ADD COLUMN paid_amount INTEGER DEFAULT 0;

-- +migrate Down

ALTER TABLE payment_records DROP COLUMN paid_amount;
