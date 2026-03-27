-- +migrate Up

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    total_cost INTEGER NOT NULL,
    split_amount INTEGER NOT NULL,
    payment_info TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_event_id ON payments(event_id);


-- +migrate Down

DROP TABLE IF EXISTS payments;

