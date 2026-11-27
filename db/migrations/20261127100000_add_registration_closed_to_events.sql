-- +migrate Up

ALTER TABLE events
    ADD COLUMN registration_closed BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down

ALTER TABLE events
    DROP COLUMN IF EXISTS registration_closed;
