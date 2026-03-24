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

// ListVenuesFilter holds filter parameters for listing venues
type ListVenuesFilter struct {
	Search string // Search in name or address
}

// ListVenuesRequest holds pagination and filter parameters
type ListVenuesRequest struct {
	Mode   PaginationMode
	Page   int    // For page-based pagination
	Limit  int    // Page size or cursor limit
	Cursor string // For cursor-based pagination (last venue ID)
	Filter ListVenuesFilter
}

// ListVenuesResponse is the paginated response
type ListVenuesResponse struct {
	Venues     []*Venue `json:"venues"`
	Total      int64    `json:"total,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
	HasMore    bool     `json:"has_more,omitempty"`
}

type VenueRepository interface {
	FindByID(ctx context.Context, id string) (*Venue, error)
	FindByCreatedBy(ctx context.Context, createdBy string) ([]*Venue, error)
	ListAll(ctx context.Context) ([]*Venue, error)
	// ListPaginated returns paginated venues with total count (filtering and pagination done in SQL)
	ListPaginated(ctx context.Context, req *ListVenuesRequest) ([]*Venue, int64, error)
	Create(ctx context.Context, venue *Venue) error
	Update(ctx context.Context, venue *Venue) error
	Delete(ctx context.Context, id string) error
}

type VenueUsecase interface {
	GetByID(ctx context.Context, id string) (*Venue, error)
	ListAll(ctx context.Context) ([]*Venue, error)
	ListPaginated(ctx context.Context, req *ListVenuesRequest) (*ListVenuesResponse, error)
	Create(ctx context.Context, userID string, req *CreateVenueRequest) (*Venue, error)
	Update(ctx context.Context, id string, req *UpdateVenueRequest) (*Venue, error)
	Delete(ctx context.Context, id string) error
}
