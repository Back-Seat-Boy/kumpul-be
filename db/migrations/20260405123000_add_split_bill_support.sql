-- +migrate Up

ALTER TABLE payments
    ADD COLUMN tax_amount INTEGER NOT NULL DEFAULT 0;

CREATE TABLE split_bill_items (
    id UUID PRIMARY KEY,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE split_bill_item_assignments (
    id UUID PRIMARY KEY,
    item_id UUID NOT NULL REFERENCES split_bill_items(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    UNIQUE (item_id, participant_id)
);

CREATE INDEX idx_split_bill_items_payment_id ON split_bill_items(payment_id);
CREATE INDEX idx_split_bill_item_assignments_item_id ON split_bill_item_assignments(item_id);
CREATE INDEX idx_split_bill_item_assignments_participant_id ON split_bill_item_assignments(participant_id);

-- +migrate Down

DROP TABLE IF EXISTS split_bill_item_assignments;
DROP TABLE IF EXISTS split_bill_items;

ALTER TABLE payments
    DROP COLUMN IF EXISTS tax_amount;
