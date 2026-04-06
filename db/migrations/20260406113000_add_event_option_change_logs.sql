-- +migrate Up

CREATE TABLE event_option_change_logs (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    event_option_id UUID NOT NULL REFERENCES event_options(id) ON DELETE CASCADE,
    edited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note TEXT,
    old_venue_id UUID NOT NULL REFERENCES venues(id),
    old_date TIMESTAMP NOT NULL,
    old_start_time TIME NOT NULL,
    old_end_time TIME NOT NULL,
    new_venue_id UUID NOT NULL REFERENCES venues(id),
    new_date TIMESTAMP NOT NULL,
    new_start_time TIME NOT NULL,
    new_end_time TIME NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_option_change_logs_event_id ON event_option_change_logs(event_id);
CREATE INDEX idx_event_option_change_logs_option_id ON event_option_change_logs(event_option_id);
CREATE INDEX idx_event_option_change_logs_created_at ON event_option_change_logs(created_at);

-- +migrate Down

DROP TABLE IF EXISTS event_option_change_logs;
