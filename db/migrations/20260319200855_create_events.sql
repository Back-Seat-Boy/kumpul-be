-- +migrate Up

CREATE TYPE event_status AS ENUM ('voting', 'confirmed', 'open', 'payment_open', 'completed');

CREATE TABLE events (
    id UUID PRIMARY KEY,
    created_by UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status event_status NOT NULL DEFAULT 'voting',
    chosen_option_id UUID,
    player_cap INTEGER,
    voting_deadline TIMESTAMP,
    share_token VARCHAR(10) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_share_token ON events(share_token);


-- +migrate Down

DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS event_status;

