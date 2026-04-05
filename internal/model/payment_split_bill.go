package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type SplitBillItemInput struct {
	Name           string   `json:"name"`
	Price          int      `json:"price"`
	ParticipantIDs []string `json:"participant_ids"`
}

type SplitBillItem struct {
	ID          string                    `json:"id" gorm:"primaryKey;type:uuid"`
	PaymentID   string                    `json:"payment_id" gorm:"type:uuid;not null;index"`
	Name        string                    `json:"name" gorm:"not null"`
	Price       int                       `json:"price" gorm:"not null"`
	CreatedAt   time.Time                 `json:"created_at"`
	Assignments []SplitBillItemAssignment `json:"assignments,omitempty" gorm:"foreignKey:ItemID"`
}

type SplitBillItemAssignment struct {
	ID            string      `json:"id" gorm:"primaryKey;type:uuid"`
	ItemID        string      `json:"item_id" gorm:"type:uuid;not null;index"`
	ParticipantID string      `json:"participant_id" gorm:"type:uuid;not null;index"`
	Participant   Participant `json:"participant,omitempty" gorm:"foreignKey:ParticipantID"`
}

type SplitBillItemAssignmentDetail struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
	Amount        int    `json:"amount"`
}

type SplitBillItemDetail struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Price       int                             `json:"price"`
	Assignments []SplitBillItemAssignmentDetail `json:"assignments"`
}

type SplitBillParticipantDetail struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
	Subtotal      int    `json:"subtotal"`
	TaxAmount     int    `json:"tax_amount"`
	Total         int    `json:"total"`
}

type SplitBillDetails struct {
	TaxAmount     int                          `json:"tax_amount"`
	ItemsSubtotal int                          `json:"items_subtotal"`
	GrandTotal    int                          `json:"grand_total"`
	Items         []SplitBillItemDetail        `json:"items"`
	Participants  []SplitBillParticipantDetail `json:"participants"`
}

type SplitBillRepository interface {
	FindByPaymentID(ctx context.Context, paymentID string) ([]*SplitBillItem, error)
	DeleteByPaymentIDWithTx(ctx context.Context, tx *gorm.DB, paymentID string) error
	CreateItemWithTx(ctx context.Context, tx *gorm.DB, item *SplitBillItem) error
	CreateAssignmentWithTx(ctx context.Context, tx *gorm.DB, assignment *SplitBillItemAssignment) error
	HasAssignmentsForParticipant(ctx context.Context, paymentID, participantID string) (bool, error)
}
