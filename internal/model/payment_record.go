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
	PaidAmount    int                 `json:"paid_amount" gorm:"default:0"`
	ProofImageURL string              `json:"proof_image_url"`
	ClaimedAt     *time.Time          `json:"claimed_at"`
	ConfirmedAt   *time.Time          `json:"confirmed_at"`
	User          User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ClaimPaymentRequest struct {
	ProofImageURL string `json:"proof_image_url"`
}

type AdjustPaymentRequest struct {
	AdjustmentAmount int    `json:"adjustment_amount" validate:"required"` // positive = paid more, negative = refunded
	ProofImageURL    string `json:"proof_image_url"`
	Note             string `json:"note"`
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
	CountConfirmedByPaymentID(ctx context.Context, paymentID string) (int64, error)
}

type ParticipantPaymentStatus struct {
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	Status          string `json:"status"`
	PaidAmount      int    `json:"paid_amount"`
	CurrentSplit    int    `json:"current_split"`
	Difference      int    `json:"difference"`
	Action          string `json:"action"`
	ActionAmount    int    `json:"action_amount"`
}

type PaymentRecordsWithSummary struct {
	Records            []*PaymentRecord         `json:"records"`
	NumParticipants    int64                    `json:"num_participants"`
	NumConfirmed       int64                    `json:"num_confirmed"`
	NumClaimed         int64                    `json:"num_claimed"`
	NumPending         int64                    `json:"num_pending"`
	TotalCollected     int                      `json:"total_collected"`
	TotalShouldCollect int                      `json:"total_should_collect"`
	Balance            int                      `json:"balance"`
	PerPersonStatus    []ParticipantPaymentStatus `json:"per_person_status"`
}

type PaymentRecordUsecase interface {
	GetByPaymentID(ctx context.Context, paymentID string) (*PaymentRecordsWithSummary, error)
	Claim(ctx context.Context, paymentID string, userID string, req *ClaimPaymentRequest) error
	Confirm(ctx context.Context, paymentID string, userID string) error
	AdjustPayment(ctx context.Context, paymentID string, userID string, requesterID string, req *AdjustPaymentRequest) error
}
