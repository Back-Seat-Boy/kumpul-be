package model

import (
	"context"
	"time"
)

type EventStatus string

const (
	EventStatusVoting      EventStatus = "voting"
	EventStatusConfirmed   EventStatus = "confirmed"
	EventStatusOpen        EventStatus = "open"
	EventStatusPaymentOpen EventStatus = "payment_open"
	EventStatusCompleted   EventStatus = "completed"
)

type Event struct {
	ID             string       `json:"id" gorm:"primaryKey;type:uuid"`
	CreatedBy      string       `json:"created_by" gorm:"type:uuid;not null"`
	Title          string       `json:"title" gorm:"not null"`
	Description    string       `json:"description"`
	Status         EventStatus  `json:"status" gorm:"not null;default:'voting'"`
	ChosenOptionID *string      `json:"chosen_option_id,omitempty" gorm:"type:uuid"`
	PlayerCap      *int         `json:"player_cap"`
	VotingDeadline *time.Time   `json:"voting_deadline"`
	ShareToken     string       `json:"share_token" gorm:"uniqueIndex;not null"`
	CreatedAt      time.Time    `json:"created_at"`
	Creator        User         `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	ChosenOption   *EventOption `json:"chosen_option,omitempty" gorm:"foreignKey:ChosenOptionID"`
}

type CreateEventRequest struct {
	Title          string     `json:"title" validate:"required"`
	Description    string     `json:"description"`
	PlayerCap      *int       `json:"player_cap"`
	VotingDeadline *time.Time `json:"voting_deadline"`
}

type UpdateEventStatusRequest struct {
	Status EventStatus `json:"status" validate:"required"`
}

type UpdateEventChosenOptionRequest struct {
	OptionID string `json:"option_id" validate:"required"`
}

type EventRepository interface {
	FindByID(ctx context.Context, id string) (*Event, error)
	FindByShareToken(ctx context.Context, token string) (*Event, error)
	FindByCreatedBy(ctx context.Context, createdBy string) ([]*Event, error)
	List(ctx context.Context) ([]*Event, error)
	Create(ctx context.Context, event *Event) error
	Update(ctx context.Context, event *Event) error
	UpdateStatus(ctx context.Context, id string, status EventStatus) error
	UpdateChosenOption(ctx context.Context, id string, optionID string) error
	Delete(ctx context.Context, id string) error
}

type EventUsecase interface {
	GetByID(ctx context.Context, id string) (*Event, error)
	GetByShareToken(ctx context.Context, token string) (*Event, error)
	List(ctx context.Context) ([]*Event, error)
	Create(ctx context.Context, userID string, req *CreateEventRequest) (*Event, error)
	UpdateStatus(ctx context.Context, id string, status EventStatus) error
	UpdateChosenOption(ctx context.Context, id string, optionID string) error
	Delete(ctx context.Context, id string) error
}
