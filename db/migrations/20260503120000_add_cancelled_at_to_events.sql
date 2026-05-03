-- +migrate Up

ALTER TABLE events
    ADD COLUMN cancelled_at TIMESTAMP;

UPDATE events
SET cancelled_at = created_at
WHERE status = 'cancelled'
  AND cancelled_at IS NULL;

CREATE INDEX idx_events_cancelled_at ON events(cancelled_at);

-- +migrate Down

DROP INDEX IF EXISTS idx_events_cancelled_at;

ALTER TABLE events
    DROP COLUMN IF EXISTS cancelled_at;
