package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type EventImage struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID   string    `json:"event_id" gorm:"type:uuid;not null;index"`
	ImageURL  string    `json:"image_url" gorm:"not null"`
	Position  int       `json:"position" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateEventImagesRequest struct {
	ImageURLs []string `json:"image_urls"`
}

type EventImageRepository interface {
	ReplaceByEventIDWithTx(ctx context.Context, tx *gorm.DB, eventID string, images []*EventImage) error
	FindByEventID(ctx context.Context, eventID string) ([]*EventImage, error)
}
