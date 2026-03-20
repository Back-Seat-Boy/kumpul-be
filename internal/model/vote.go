package model

import (
	"context"
	"time"
)

type Vote struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid"`
	EventOptionID string    `json:"event_option_id" gorm:"type:uuid;not null"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null"`
	CreatedAt     time.Time `json:"created_at"`
}

type CastVoteRequest struct {
	EventOptionID string `json:"event_option_id" validate:"required"`
}

type VoteRepository interface {
	FindByEventOptionIDAndUserID(ctx context.Context, eventOptionID, userID string) (*Vote, error)
	Create(ctx context.Context, vote *Vote) error
	Delete(ctx context.Context, id string) error
}

type VoteUsecase interface {
	CastVote(ctx context.Context, userID string, req *CastVoteRequest) error
	RemoveVote(ctx context.Context, eventOptionID string, userID string) error
}
