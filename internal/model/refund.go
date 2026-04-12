package model

import (
	"context"
	"time"
)

type RefundStatus string

const (
	RefundStatusPendingInfo RefundStatus = "pending_info"
	RefundStatusReadyToSend RefundStatus = "ready_to_send"
	RefundStatusSent        RefundStatus = "sent"
	RefundStatusReceived    RefundStatus = "received"
	RefundStatusCancelled   RefundStatus = "cancelled"
)

type Refund struct {
	ID                       string       `json:"id" gorm:"primaryKey;type:uuid"`
	EventID                  string       `json:"event_id" gorm:"type:uuid;not null;index"`
	PaymentID                string       `json:"payment_id" gorm:"type:uuid;not null;index"`
	RemovedParticipantID     string       `json:"removed_participant_id" gorm:"type:uuid;not null"`
	UserID                   *string      `json:"user_id,omitempty" gorm:"type:uuid;index"`
	DisplayName              string       `json:"display_name" gorm:"not null"`
	Amount                   int          `json:"amount" gorm:"not null"`
	Status                   RefundStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending_info'"`
	RecipientPaymentMethodID *string      `json:"recipient_payment_method_id,omitempty" gorm:"type:uuid"`
	RecipientPaymentInfo     string       `json:"recipient_payment_info,omitempty"`
	RecipientPaymentImageURL string       `json:"recipient_payment_image_url,omitempty"`
	RecipientNote            string       `json:"recipient_note,omitempty"`
	SentProofImageURL        string       `json:"sent_proof_image_url,omitempty"`
	SentNote                 string       `json:"sent_note,omitempty"`
	SentAt                   *time.Time   `json:"sent_at,omitempty"`
	ReceivedAt               *time.Time   `json:"received_at,omitempty"`
	CreatedAt                time.Time    `json:"created_at"`
	UpdatedAt                time.Time    `json:"updated_at"`
}

type UpdateRefundDestinationRequest struct {
	PaymentMethodID string `json:"payment_method_id"`
	PaymentInfo     string `json:"payment_info"`
	PaymentImageURL string `json:"payment_image_url"`
	Note            string `json:"note"`
}

type SendRefundRequest struct {
	ProofImageURL string `json:"proof_image_url"`
	Note          string `json:"note"`
}

type RefundRepository interface {
	FindByID(ctx context.Context, id string) (*Refund, error)
	FindByEventID(ctx context.Context, eventID string) ([]*Refund, error)
	FindByUserID(ctx context.Context, userID string) ([]*Refund, error)
	Create(ctx context.Context, refund *Refund) error
	Update(ctx context.Context, refund *Refund) error
}

type RefundUsecase interface {
	CreateForRemovedParticipant(ctx context.Context, event *Event, payment *Payment, participant *Participant, record *PaymentRecord) (*Refund, error)
	ListByEvent(ctx context.Context, eventID string, requesterID string) ([]*Refund, error)
	ListByUserID(ctx context.Context, userID string) ([]*Refund, error)
	UpdateDestination(ctx context.Context, id string, userID string, req *UpdateRefundDestinationRequest) (*Refund, error)
	MarkSent(ctx context.Context, id string, requesterID string, req *SendRefundRequest) (*Refund, error)
	ConfirmReceipt(ctx context.Context, id string, userID string) (*Refund, error)
}
