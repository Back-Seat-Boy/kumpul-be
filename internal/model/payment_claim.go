package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PaymentClaimStatus string

const (
	PaymentClaimStatusClaimed   PaymentClaimStatus = "claimed"
	PaymentClaimStatusConfirmed PaymentClaimStatus = "confirmed"
)

type PaymentClaim struct {
	ID              string             `json:"id" gorm:"primaryKey;type:uuid"`
	PaymentRecordID string             `json:"payment_record_id" gorm:"type:uuid;not null"`
	ParticipantID   string             `json:"participant_id" gorm:"type:uuid;not null"`
	ClaimedAmount   int                `json:"claimed_amount" gorm:"not null;default:0"`
	ProofImageURL   string             `json:"proof_image_url"`
	Note            string             `json:"note"`
	Status          PaymentClaimStatus `json:"status" gorm:"not null;default:'claimed'"`
	ClaimedAt       time.Time          `json:"claimed_at" gorm:"not null"`
	ConfirmedAt     *time.Time         `json:"confirmed_at"`
}

type PaymentClaimRepository interface {
	Create(ctx context.Context, claim *PaymentClaim) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, claim *PaymentClaim) error
	Update(ctx context.Context, claim *PaymentClaim) error
	FindLatestClaimedByPaymentRecordID(ctx context.Context, paymentRecordID string) (*PaymentClaim, error)
	FindByPaymentRecordID(ctx context.Context, paymentRecordID string) ([]*PaymentClaim, error)
}
