package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type EventOption struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID   string    `json:"event_id" gorm:"type:uuid;not null"`
	VenueID   string    `json:"venue_id" gorm:"type:uuid;not null"`
	Date      time.Time `json:"date"`
	StartTime string    `json:"start_time" gorm:"type:time"`
	EndTime   string    `json:"end_time" gorm:"type:time"`
	Event     Event     `json:"event,omitempty" gorm:"foreignKey:EventID"`
	Venue     Venue     `json:"venue,omitempty" gorm:"foreignKey:VenueID"`
}

type VoterInfo struct {
	EventOptionID string `json:"event_option_id"`
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	AvatarURL     string `json:"avatar_url"`
}

type EventOptionWithVoteCount struct {
	EventOption
	VoteCount int64       `json:"vote_count"`
	HasVoted  bool        `json:"has_voted"`
	Voters    []VoterInfo `json:"voters,omitempty"`
}

type CreateEventOptionRequest struct {
	VenueID   string    `json:"venue_id"`
	Date      time.Time `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
}

type UpdateEventOptionRequest struct {
	VenueID   string    `json:"venue_id" validate:"required"`
	Date      time.Time `json:"date" validate:"required"`
	StartTime string    `json:"start_time" validate:"required"`
	EndTime   string    `json:"end_time" validate:"required"`
	Note      string    `json:"note"`
}

type EventOptionRepository interface {
	FindByID(ctx context.Context, id string) (*EventOption, error)
	FindByEventID(ctx context.Context, eventID string) ([]*EventOption, error)
	FindByEventIDWithVoteCount(ctx context.Context, eventID string, userID *string) ([]*EventOptionWithVoteCount, error)
	FindVotersByOptionID(ctx context.Context, optionID string) ([]VoterInfo, error)
	FindVotersByOptionIDs(ctx context.Context, optionID []string) ([]VoterInfo, error)
	Create(ctx context.Context, option *EventOption) error
	BulkCreateWithTx(ctx context.Context, tx *gorm.DB, options []*EventOption) error
	UpdateScheduleWithTx(ctx context.Context, tx *gorm.DB, optionID string, venueID string, date time.Time, startTime string, endTime string) error
	UpdateWithTx(ctx context.Context, tx *gorm.DB, optionID string, venueID string, date time.Time, startTime string, endTime string) error
	Delete(ctx context.Context, id string) error
}

type EventOptionUsecase interface {
	GetByID(ctx context.Context, id string) (*EventOption, error)
	ListByEvent(ctx context.Context, eventID string, userID *string) ([]*EventOptionWithVoteCount, error)
	ListByEventWithVoters(ctx context.Context, eventID string, userID *string) ([]*EventOptionWithVoteCount, error)
	ListChangeLogs(ctx context.Context, eventID string, userID string) ([]*EventOptionChangeLog, error)
	Create(ctx context.Context, eventID string, req *CreateEventOptionRequest) (*EventOption, error)
	Update(ctx context.Context, eventID string, optionID string, userID string, req *UpdateEventOptionRequest) error
	Delete(ctx context.Context, id string) error
}
