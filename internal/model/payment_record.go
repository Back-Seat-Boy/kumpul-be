package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PaymentRecordStatus string

const (
	PaymentRecordStatusPending   PaymentRecordStatus = "pending"
	PaymentRecordStatusClaimed   PaymentRecordStatus = "claimed"
	PaymentRecordStatusConfirmed PaymentRecordStatus = "confirmed"
)

type PaymentRecord struct {
	ID            string              `json:"id" gorm:"primaryKey;type:uuid"`
	PaymentID     string              `json:"payment_id" gorm:"type:uuid;not null"`
	UserID        string              `json:"user_id" gorm:"type:uuid;not null"`
	Status        PaymentRecordStatus `json:"status" gorm:"not null;default:'pending'"`
	ProofImageURL string              `json:"proof_image_url"`
	ClaimedAt     *time.Time          `json:"claimed_at"`
	ConfirmedAt   *time.Time          `json:"confirmed_at"`
	User          User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ClaimPaymentRequest struct {
	ProofImageURL string `json:"proof_image_url"`
}

type PaymentRecordRepository interface {
	FindByPaymentID(ctx context.Context, paymentID string) ([]*PaymentRecord, error)
	FindByPaymentIDAndUserID(ctx context.Context, paymentID, userID string) (*PaymentRecord, error)
	Create(ctx context.Context, record *PaymentRecord) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, record *PaymentRecord) error
	Update(ctx context.Context, record *PaymentRecord) error
	DeleteByPaymentIDAndUserID(ctx context.Context, paymentID, userID string) error
	DeleteByPaymentIDAndUserIDWithTx(ctx context.Context, tx *gorm.DB, paymentID, userID string) error
	UpdateSplitAmountByPaymentID(ctx context.Context, paymentID string, splitAmount int) error
}

type PaymentRecordUsecase interface {
	GetByPaymentID(ctx context.Context, paymentID string) ([]*PaymentRecord, error)
	Claim(ctx context.Context, paymentID string, userID string, req *ClaimPaymentRequest) error
	Confirm(ctx context.Context, paymentID string, userID string) error
}
