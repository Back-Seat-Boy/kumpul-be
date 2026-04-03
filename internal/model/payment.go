package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PaymentType string

const (
	PaymentTypeTotal     PaymentType = "total"
	PaymentTypePerPerson PaymentType = "per_person"
)

type Payment struct {
	ID          string      `json:"id" gorm:"primaryKey;type:uuid"`
	EventID     string      `json:"event_id" gorm:"type:uuid;uniqueIndex;not null"`
	TotalCost   int         `json:"total_cost" gorm:"not null"`
	BaseSplit   int         `json:"base_split" gorm:"column:base_split;not null"`
	Type        PaymentType `json:"type" gorm:"type:varchar(20);not null;default:'total'"`
	PaymentInfo string      `json:"payment_info" gorm:"not null"`
	CreatedAt   time.Time   `json:"created_at"`
}

type CreatePaymentRequest struct {
	Type            string `json:"type"`
	TotalCost       int    `json:"total_cost"`
	PerPersonAmount int    `json:"per_person_amount"`
	PaymentInfo     string `json:"payment_info" validate:"required"`
}

type UpdatePaymentRequest struct {
	PaymentInfo string `json:"payment_info" validate:"required"`
}

type UpdatePaymentConfigRequest struct {
	Type            string `json:"type"`
	TotalCost       int    `json:"total_cost"`
	PerPersonAmount int    `json:"per_person_amount"`
}

type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByEventID(ctx context.Context, eventID string) (*Payment, error)
	Create(ctx context.Context, payment *Payment) error
	UpdateBaseSplit(ctx context.Context, id string, baseSplit int) error
	UpdateBaseSplitWithTx(ctx context.Context, tx *gorm.DB, id string, baseSplit int) error
	UpdateTotals(ctx context.Context, id string, totalCost, baseSplit int) error
	UpdateTotalsWithTx(ctx context.Context, tx *gorm.DB, id string, totalCost, baseSplit int) error
	UpdateConfigWithTx(ctx context.Context, tx *gorm.DB, id string, paymentType PaymentType, totalCost, baseSplit int) error
	UpdatePaymentInfo(ctx context.Context, id string, paymentInfo string) error
}

type PaymentUsecase interface {
	GetByEventID(ctx context.Context, eventID string) (*Payment, error)
	Create(ctx context.Context, eventID string, req *CreatePaymentRequest) (*Payment, error)
	RecalculateSplitAmount(ctx context.Context, eventID string) error
	UpdatePaymentInfo(ctx context.Context, eventID string, requesterID string, req *UpdatePaymentRequest) (*Payment, error)
	UpdatePaymentConfig(ctx context.Context, eventID string, requesterID string, req *UpdatePaymentConfigRequest) (*Payment, error)
	ChargeAll(ctx context.Context, eventID string, requesterID string, req *ChargeAllRequest) error
}
