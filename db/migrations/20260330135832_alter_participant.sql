-- +migrate Up

ALTER TABLE participants
    ALTER COLUMN user_id DROP NOT NULL,
    ADD COLUMN guest_name VARCHAR(255),
    ADD COLUMN added_by UUID;

-- +migrate Down

ALTER TABLE participants
    ALTER COLUMN user_id SET NOT NULL,
    DROP COLUMN IF EXISTS guest_name,
    DROP COLUMN IF EXISTS added_by;