package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID     string    `json:"event_id" gorm:"type:uuid;uniqueIndex;not null"`
	TotalCost   int       `json:"total_cost" gorm:"not null"`
	BaseSplit   int       `json:"base_split" gorm:"column:base_split;not null"`
	PaymentInfo string    `json:"payment_info" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreatePaymentRequest struct {
	TotalCost   int    `json:"total_cost" validate:"required,gt=0"`
	PaymentInfo string `json:"payment_info" validate:"required"`
}

type UpdatePaymentRequest struct {
	PaymentInfo string `json:"payment_info" validate:"required"`
}

type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByEventID(ctx context.Context, eventID string) (*Payment, error)
	Create(ctx context.Context, payment *Payment) error
	UpdateBaseSplit(ctx context.Context, id string, baseSplit int) error
	UpdateBaseSplitWithTx(ctx context.Context, tx *gorm.DB, id string, baseSplit int) error
	UpdatePaymentInfo(ctx context.Context, id string, paymentInfo string) error
}

type PaymentUsecase interface {
	GetByEventID(ctx context.Context, eventID string) (*Payment, error)
	Create(ctx context.Context, eventID string, req *CreatePaymentRequest) (*Payment, error)
	RecalculateSplitAmount(ctx context.Context, eventID string) error
	UpdatePaymentInfo(ctx context.Context, eventID string, requesterID string, req *UpdatePaymentRequest) (*Payment, error)
	ChargeAll(ctx context.Context, eventID string, requesterID string, req *ChargeAllRequest) error
}
