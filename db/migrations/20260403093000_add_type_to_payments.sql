-- +migrate Up

ALTER TABLE payments
    ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'total';

UPDATE payments
SET type = 'total'
WHERE type IS NULL OR type = '';

-- +migrate Down

ALTER TABLE payments
    DROP COLUMN IF EXISTS type;
