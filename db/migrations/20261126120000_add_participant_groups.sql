-- +migrate Up

CREATE TABLE participant_groups (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_participant_groups_event_name UNIQUE (event_id, name)
);

CREATE INDEX idx_participant_groups_event_id ON participant_groups(event_id);

ALTER TABLE participants
    ADD COLUMN group_id UUID NULL REFERENCES participant_groups(id) ON DELETE SET NULL;

CREATE INDEX idx_participants_group_id ON participants(group_id);

ALTER TABLE events
    ADD COLUMN allow_group_self_select BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down

ALTER TABLE events
    DROP COLUMN IF EXISTS allow_group_self_select;

DROP INDEX IF EXISTS idx_participants_group_id;
ALTER TABLE participants
    DROP COLUMN IF EXISTS group_id;

DROP INDEX IF EXISTS idx_participant_groups_event_id;
DROP TABLE IF EXISTS participant_groups;