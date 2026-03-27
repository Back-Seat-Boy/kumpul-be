-- +migrate Up

CREATE TABLE event_options (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    venue_id UUID NOT NULL REFERENCES venues(id),
    date DATE,
    start_time TIME,
    end_time TIME
);

CREATE INDEX idx_event_options_event_id ON event_options(event_id);
CREATE INDEX idx_event_options_venue_id ON event_options(venue_id);

ALTER TABLE events ADD CONSTRAINT fk_events_chosen_option 
    FOREIGN KEY (chosen_option_id) REFERENCES event_options(id);


-- +migrate Down

ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_chosen_option;
DROP TABLE IF EXISTS event_options;

