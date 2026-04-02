-- +migrate Up

CREATE TYPE payment_claim_status AS ENUM ('claimed', 'confirmed');

CREATE TABLE payment_claims (
    id UUID PRIMARY KEY,
    payment_record_id UUID NOT NULL REFERENCES payment_records(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    claimed_amount INTEGER NOT NULL DEFAULT 0,
    proof_image_url TEXT,
    note TEXT,
    status payment_claim_status NOT NULL DEFAULT 'claimed',
    claimed_at TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP
);

CREATE INDEX idx_payment_claims_payment_record_id ON payment_claims(payment_record_id);
CREATE INDEX idx_payment_claims_participant_id ON payment_claims(participant_id);

INSERT INTO payment_claims (
    id,
    payment_record_id,
    participant_id,
    claimed_amount,
    proof_image_url,
    note,
    status,
    claimed_at,
    confirmed_at
)
SELECT
    gen_random_uuid(),
    pr.id,
    pr.participant_id,
    CASE
        WHEN pr.status = 'confirmed' THEN pr.paid_amount
        ELSE GREATEST(pr.amount - pr.paid_amount, 0)
    END,
    pr.proof_image_url,
    pr.note,
    CASE
        WHEN pr.status = 'confirmed' THEN 'confirmed'::payment_claim_status
        ELSE 'claimed'::payment_claim_status
    END,
    COALESCE(pr.claimed_at, pr.confirmed_at, NOW()),
    pr.confirmed_at
FROM payment_records pr
WHERE pr.status IN ('claimed', 'confirmed')
  AND (
      pr.paid_amount > 0
      OR pr.proof_image_url IS NOT NULL
      OR pr.claimed_at IS NOT NULL
      OR pr.confirmed_at IS NOT NULL
  );

-- +migrate Down

DROP TABLE IF EXISTS payment_claims;
DROP TYPE IF EXISTS payment_claim_status;
