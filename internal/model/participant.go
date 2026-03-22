package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Participant struct {
	ID       string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventID  string    `json:"event_id" gorm:"type:uuid;not null"`
	UserID   string    `json:"user_id" gorm:"type:uuid;not null"`
	JoinedAt time.Time `json:"joined_at"`
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ParticipantRepository interface {
	FindByEventID(ctx context.Context, eventID string) ([]*Participant, error)
	FindByEventIDAndUserID(ctx context.Context, eventID, userID string) (*Participant, error)
	Create(ctx context.Context, participant *Participant) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, participant *Participant) error
	Delete(ctx context.Context, id string) error
	CountByEventID(ctx context.Context, eventID string) (int64, error)
}

type ParticipantUsecase interface {
	ListByEvent(ctx context.Context, eventID string) ([]*Participant, error)
	Join(ctx context.Context, eventID string, userID string) error
	Leave(ctx context.Context, eventID string, userID string) error
	GetParticipantCount(ctx context.Context, eventID string) (int64, error)
	HandlePaymentOnJoin(ctx context.Context, eventID string, userID string) error
	HandlePaymentOnLeave(ctx context.Context, eventID string, userID string) error
}
