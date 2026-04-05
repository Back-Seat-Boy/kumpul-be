-- +migrate Up

ALTER TABLE events
    ADD COLUMN visibility VARCHAR(20) NOT NULL DEFAULT 'public';

CREATE INDEX IF NOT EXISTS idx_events_visibility ON events (visibility);

UPDATE events
SET visibility = 'public'
WHERE visibility IS NULL OR visibility = '';

-- +migrate Down

DROP INDEX IF EXISTS idx_events_visibility;

ALTER TABLE events
    DROP COLUMN IF EXISTS visibility;
