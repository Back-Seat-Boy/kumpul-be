package model

import (
	"context"
	"time"
)

type Venue struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid"`
	CreatedBy      string    `json:"created_by" gorm:"type:uuid;not null"`
	Name           string    `json:"name" gorm:"not null"`
	Address        string    `json:"address"`
	WhatsappNumber string    `json:"whatsapp_number"`
	PricePerHour   int       `json:"price_per_hour"`
	CourtCount     int       `json:"court_count"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	Creator        User      `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

type CreateVenueRequest struct {
	Name           string `json:"name" validate:"required"`
	Address        string `json:"address"`
	WhatsappNumber string `json:"whatsapp_number"`
	PricePerHour   int    `json:"price_per_hour"`
	CourtCount     int    `json:"court_count"`
	Notes          string `json:"notes"`
}

type UpdateVenueRequest struct {
	Name           string `json:"name" validate:"required"`
	Address        string `json:"address"`
	WhatsappNumber string `json:"whatsapp_number"`
	PricePerHour   int    `json:"price_per_hour"`
	CourtCount     int    `json:"court_count"`
	Notes          string `json:"notes"`
}

type VenueRepository interface {
	FindByID(ctx context.Context, id string) (*Venue, error)
	FindByCreatedBy(ctx context.Context, createdBy string) ([]*Venue, error)
	ListAll(ctx context.Context) ([]*Venue, error)
	Create(ctx context.Context, venue *Venue) error
	Update(ctx context.Context, venue *Venue) error
	Delete(ctx context.Context, id string) error
}

type VenueUsecase interface {
	GetByID(ctx context.Context, id string) (*Venue, error)
	ListAll(ctx context.Context) ([]*Venue, error)
	Create(ctx context.Context, userID string, req *CreateVenueRequest) (*Venue, error)
	Update(ctx context.Context, id string, req *UpdateVenueRequest) (*Venue, error)
	Delete(ctx context.Context, id string) error
}
