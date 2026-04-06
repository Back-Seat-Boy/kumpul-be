package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type EventOptionChangeLog struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID       string    `json:"event_id" gorm:"type:uuid;not null"`
	EventOptionID string    `json:"event_option_id" gorm:"type:uuid;not null"`
	EditedBy      string    `json:"edited_by" gorm:"type:uuid;not null"`
	Note          string    `json:"note" gorm:"type:text"`
	OldVenueID    string    `json:"old_venue_id" gorm:"type:uuid;not null"`
	OldDate       time.Time `json:"old_date" gorm:"not null"`
	OldStartTime  string    `json:"old_start_time" gorm:"type:time;not null"`
	OldEndTime    string    `json:"old_end_time" gorm:"type:time;not null"`
	NewVenueID    string    `json:"new_venue_id" gorm:"type:uuid;not null"`
	NewDate       time.Time `json:"new_date" gorm:"not null"`
	NewStartTime  string    `json:"new_start_time" gorm:"type:time;not null"`
	NewEndTime    string    `json:"new_end_time" gorm:"type:time;not null"`
	CreatedAt     time.Time `json:"created_at"`
	Editor        User      `json:"editor" gorm:"foreignKey:EditedBy"`
	OldVenue      Venue     `json:"old_venue" gorm:"foreignKey:OldVenueID"`
	NewVenue      Venue     `json:"new_venue" gorm:"foreignKey:NewVenueID"`
}

type EventOptionChangeLogRepository interface {
	CreateWithTx(ctx context.Context, tx *gorm.DB, log *EventOptionChangeLog) error
	FindByEventID(ctx context.Context, eventID string) ([]*EventOptionChangeLog, error)
}
