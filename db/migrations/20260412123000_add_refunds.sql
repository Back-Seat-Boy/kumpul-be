-- +migrate Up

CREATE TABLE refunds (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    removed_participant_id UUID NOT NULL,
    user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    display_name VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_info',
    recipient_payment_method_id UUID NULL REFERENCES payment_methods(id) ON DELETE SET NULL,
    recipient_payment_info TEXT,
    recipient_payment_image_url TEXT,
    recipient_note TEXT,
    sent_proof_image_url TEXT,
    sent_note TEXT,
    sent_at TIMESTAMP NULL,
    received_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refunds_event_id ON refunds(event_id);
CREATE INDEX idx_refunds_payment_id ON refunds(payment_id);
CREATE INDEX idx_refunds_user_id ON refunds(user_id);
CREATE INDEX idx_refunds_status ON refunds(status);

-- +migrate Down

DROP INDEX IF EXISTS idx_refunds_status;
DROP INDEX IF EXISTS idx_refunds_user_id;
DROP INDEX IF EXISTS idx_refunds_payment_id;
DROP INDEX IF EXISTS idx_refunds_event_id;
DROP TABLE IF EXISTS refunds;
