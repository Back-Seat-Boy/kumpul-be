-- +migrate Up

CREATE TYPE payment_record_status AS ENUM ('pending', 'claimed', 'confirmed');

CREATE TABLE payment_records (
    id UUID PRIMARY KEY,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    status payment_record_status NOT NULL DEFAULT 'pending',
    proof_image_url TEXT,
    claimed_at TIMESTAMP,
    confirmed_at TIMESTAMP
);

CREATE INDEX idx_payment_records_payment_id ON payment_records(payment_id);
CREATE INDEX idx_payment_records_user_id ON payment_records(user_id);


-- +migrate Down

DROP TABLE IF EXISTS payment_records;
DROP TYPE IF EXISTS payment_record_status;

