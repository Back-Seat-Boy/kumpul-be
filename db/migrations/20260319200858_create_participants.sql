-- +migrate Up

CREATE TABLE participants (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, user_id)
);

CREATE INDEX idx_participants_event_id ON participants(event_id);
CREATE INDEX idx_participants_user_id ON participants(user_id);


-- +migrate Down

DROP TABLE IF EXISTS participants;

