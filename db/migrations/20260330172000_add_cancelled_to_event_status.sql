-- +migrate Up

ALTER TYPE event_status ADD VALUE IF NOT EXISTS 'cancelled';


-- +migrate Down

-- PostgreSQL enum value removal is not reversible in-place.
