-- +migrate Up

CREATE TABLE venues (
    id UUID PRIMARY KEY,
    created_by UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    address TEXT,
    whatsapp_number VARCHAR(20),
    price_per_hour INTEGER,
    court_count INTEGER,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_venues_created_by ON venues(created_by);


-- +migrate Down

DROP TABLE IF EXISTS venues;

