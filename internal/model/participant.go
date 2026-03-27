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

type RemoveParticipantImpact struct {
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	PaidAmount   int    `json:"paid_amount"`
	NewSplit     int    `json:"new_split"`
	Difference   int    `json:"difference"`
	Action       string `json:"action"`
	ActionAmount int    `json:"action_amount"`
}

type RemoveParticipantResult struct {
	RemovedParticipantID string                    `json:"removed_participant_id"`
	OldSplitAmount       int                       `json:"old_split_amount"`
	NewSplitAmount       int                       `json:"new_split_amount"`
	NumPaidParticipants  int64                     `json:"num_paid_participants"`
	NumRemaining         int64                     `json:"num_remaining_participants"`
	TotalCollected       int                       `json:"total_collected"`
	TotalShouldCollect   int                       `json:"total_should_collect"`
	Difference           int                       `json:"difference"`
	Impacts              []RemoveParticipantImpact `json:"impacts"`
}

type ParticipantUsecase interface {
	ListByEvent(ctx context.Context, eventID string) ([]*Participant, error)
	Join(ctx context.Context, eventID string, userID string) error
	Leave(ctx context.Context, eventID string, userID string) error
	RemoveParticipant(ctx context.Context, eventID string, participantUserID string, requesterUserID string) (*RemoveParticipantResult, error)
	GetParticipantCount(ctx context.Context, eventID string) (int64, error)
	HandlePaymentOnJoin(ctx context.Context, eventID string, userID string) error
	HandlePaymentOnLeave(ctx context.Context, eventID string, userID string) error
}
