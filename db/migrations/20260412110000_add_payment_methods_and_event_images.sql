-- +migrate Up

CREATE TABLE payment_methods (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(255) NOT NULL,
    payment_info TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_methods_user_id ON payment_methods(user_id);

ALTER TABLE payments
    ADD COLUMN payment_method_id UUID NULL REFERENCES payment_methods(id) ON DELETE SET NULL,
    ADD COLUMN payment_image_url TEXT;

CREATE INDEX idx_payments_payment_method_id ON payments(payment_method_id);

CREATE TABLE event_images (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_event_images_position CHECK (position BETWEEN 1 AND 3),
    CONSTRAINT uq_event_images_event_position UNIQUE (event_id, position)
);

CREATE INDEX idx_event_images_event_id ON event_images(event_id);

-- +migrate Down

DROP INDEX IF EXISTS idx_event_images_event_id;
DROP TABLE IF EXISTS event_images;

DROP INDEX IF EXISTS idx_payments_payment_method_id;
ALTER TABLE payments
    DROP COLUMN IF EXISTS payment_image_url,
    DROP COLUMN IF EXISTS payment_method_id;

DROP INDEX IF EXISTS idx_payment_methods_user_id;
DROP TABLE IF EXISTS payment_methods;
