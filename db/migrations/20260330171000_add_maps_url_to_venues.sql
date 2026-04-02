-- +migrate Up

ALTER TABLE venues
    ADD COLUMN maps_url TEXT;


-- +migrate Down

ALTER TABLE venues
    DROP COLUMN IF EXISTS maps_url;
