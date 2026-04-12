package model

import (
	"context"
	"time"
)

type PaymentMethod struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Label       string    `json:"label" gorm:"not null"`
	PaymentInfo string    `json:"payment_info" gorm:"not null"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePaymentMethodRequest struct {
	Label       string `json:"label" validate:"required"`
	PaymentInfo string `json:"payment_info" validate:"required"`
	ImageURL    string `json:"image_url"`
}

type UpdatePaymentMethodRequest struct {
	Label       string `json:"label" validate:"required"`
	PaymentInfo string `json:"payment_info" validate:"required"`
	ImageURL    string `json:"image_url"`
}

type PaymentMethodRepository interface {
	FindByID(ctx context.Context, id string) (*PaymentMethod, error)
	FindByIDAndUserID(ctx context.Context, id, userID string) (*PaymentMethod, error)
	ListByUserID(ctx context.Context, userID string) ([]*PaymentMethod, error)
	Create(ctx context.Context, paymentMethod *PaymentMethod) error
	Update(ctx context.Context, paymentMethod *PaymentMethod) error
	Delete(ctx context.Context, id string) error
}

type PaymentMethodUsecase interface {
	ListByUserID(ctx context.Context, userID string) ([]*PaymentMethod, error)
	Create(ctx context.Context, userID string, req *CreatePaymentMethodRequest) (*PaymentMethod, error)
	Update(ctx context.Context, id string, userID string, req *UpdatePaymentMethodRequest) (*PaymentMethod, error)
	Delete(ctx context.Context, id string, userID string) error
}
