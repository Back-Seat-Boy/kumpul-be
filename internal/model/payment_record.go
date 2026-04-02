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
	ParticipantID string              `json:"participant_id" gorm:"type:uuid;not null"`
	Status        PaymentRecordStatus `json:"status" gorm:"not null;default:'pending'"`
	Amount        int                 `json:"amount" gorm:"not null;default:0"`
	PaidAmount    int                 `json:"paid_amount" gorm:"default:0"`
	Note          string              `json:"note"`
	ProofImageURL string              `json:"-" gorm:"column:proof_image_url"`
	ClaimedAt     *time.Time          `json:"-" gorm:"column:claimed_at"`
	ConfirmedAt   *time.Time          `json:"-" gorm:"column:confirmed_at"`
	Participant   Participant         `json:"participant,omitempty" gorm:"foreignKey:ParticipantID"`
	Claims        []PaymentClaim      `json:"claims,omitempty" gorm:"foreignKey:PaymentRecordID"`
}

type ClaimPaymentRequest struct {
	ProofImageURL string `json:"proof_image_url"`
}

type ConfirmPaymentRequest struct {
	Amount        *int   `json:"amount"`
	ProofImageURL string `json:"proof_image_url"`
	Note          string `json:"note"`
}

type AdjustPaymentRequest struct {
	Amount        int    `json:"amount" validate:"required,gte=0"`
	ProofImageURL string `json:"proof_image_url"`
	Note          string `json:"note"`
}

type ChargeAllRequest struct {
	Amount int    `json:"amount" validate:"required,gt=0"`
	Note   string `json:"note"`
}

type PaymentRecordRepository interface {
	FindByPaymentID(ctx context.Context, paymentID string) ([]*PaymentRecord, error)
	FindByPaymentIDAndParticipantID(ctx context.Context, paymentID, participantID string) (*PaymentRecord, error)
	Create(ctx context.Context, record *PaymentRecord) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, record *PaymentRecord) error
	Update(ctx context.Context, record *PaymentRecord) error
	DeleteByPaymentIDAndParticipantID(ctx context.Context, paymentID, participantID string) error
	DeleteByPaymentIDAndParticipantIDWithTx(ctx context.Context, tx *gorm.DB, paymentID, participantID string) error
	UpdateSplitAmountByPaymentID(ctx context.Context, paymentID string, splitAmount int) error
	UpdateSplitAmountByPaymentIDWithTx(ctx context.Context, tx *gorm.DB, paymentID string, splitAmount int) error
	CountConfirmedByPaymentID(ctx context.Context, paymentID string) (int64, error)
}

type ParticipantPaymentStatus struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
	Status        string `json:"status"`
	Amount        int    `json:"amount"`
	PaidAmount    int    `json:"paid_amount"`
	Difference    int    `json:"difference"`
	Action        string `json:"action"`
	ActionAmount  int    `json:"action_amount"`
}

type PaymentRecordsWithSummary struct {
	Records            []*PaymentRecord           `json:"records"`
	NumParticipants    int64                      `json:"num_participants"`
	NumConfirmed       int64                      `json:"num_confirmed"`
	NumClaimed         int64                      `json:"num_claimed"`
	NumPending         int64                      `json:"num_pending"`
	TotalCollected     int                        `json:"total_collected"`
	TotalShouldCollect int                        `json:"total_should_collect"`
	Balance            int                        `json:"balance"`
	PerPersonStatus    []ParticipantPaymentStatus `json:"per_person_status"`
}

type PaymentRecordUsecase interface {
	GetByPaymentID(ctx context.Context, paymentID string) (*PaymentRecordsWithSummary, error)
	Claim(ctx context.Context, paymentID string, userID string, req *ClaimPaymentRequest) error
	Confirm(ctx context.Context, paymentID string, participantID string, req *ConfirmPaymentRequest) error
	AdjustPayment(ctx context.Context, paymentID string, participantID string, requesterID string, req *AdjustPaymentRequest) error
	ChargeAll(ctx context.Context, paymentID string, requesterID string, req *ChargeAllRequest) error
}
