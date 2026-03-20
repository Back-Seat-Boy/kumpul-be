package model

import (
	"context"
	"time"
)

type Payment struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID     string    `json:"event_id" gorm:"type:uuid;uniqueIndex;not null"`
	TotalCost   int       `json:"total_cost" gorm:"not null"`
	SplitAmount int       `json:"split_amount" gorm:"not null"`
	PaymentInfo string    `json:"payment_info" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreatePaymentRequest struct {
	TotalCost   int    `json:"total_cost" validate:"required,gt=0"`
	PaymentInfo string `json:"payment_info" validate:"required"`
}

type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByEventID(ctx context.Context, eventID string) (*Payment, error)
	Create(ctx context.Context, payment *Payment) error
}

type PaymentUsecase interface {
	GetByEventID(ctx context.Context, eventID string) (*Payment, error)
	Create(ctx context.Context, eventID string, req *CreatePaymentRequest) (*Payment, error)
}
