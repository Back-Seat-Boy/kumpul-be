package model

import (
	"context"
	"time"
)

// Session represents a user session stored in Redis
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *Session, ttl time.Duration) error
	FindByID(ctx context.Context, sessionID string) (*Session, error)
	Delete(ctx context.Context, sessionID string) error
}

type SessionUsecase interface {
	Create(ctx context.Context, user *User) (*Session, error)
	GetByID(ctx context.Context, sessionID string) (*Session, error)
	Delete(ctx context.Context, sessionID string) error
}

// Context key type for session
type contextKey string

const ContextKeySession = contextKey("session")
const ContextKeyUser = contextKey("user")
