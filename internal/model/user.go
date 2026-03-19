package model

import (
	"context"
	"time"
)

type User struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key"`
	GoogleID       string    `json:"-" gorm:"uniqueIndex;not null"`
	Name           string    `json:"name" gorm:"not null"`
	Email          string    `json:"email" gorm:"uniqueIndex;not null"`
	EmailVerified  bool      `json:"email_verified" gorm:"not null;default:false"`
	WhatsappNumber string    `json:"whatsapp_number,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	Provider       string    `json:"-" gorm:"not null;default:'google'"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null"`
}

type UpdateUserInput struct {
	Name           string `json:"name" validate:"required"`
	WhatsappNumber string `json:"whatsapp_number" validate:"required"`
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*User, error)
}

type UserUsecase interface {
	GetByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, id string, req *UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*User, error)
}
