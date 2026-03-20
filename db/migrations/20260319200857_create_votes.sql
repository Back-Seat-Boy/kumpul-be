-- +migrate Up

CREATE TABLE votes (
    id UUID PRIMARY KEY,
    event_option_id UUID NOT NULL REFERENCES event_options(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_option_id, user_id)
);

CREATE INDEX idx_votes_event_option_id ON votes(event_option_id);
CREATE INDEX idx_votes_user_id ON votes(user_id);


-- +migrate Down

DROP TABLE IF EXISTS votes;

