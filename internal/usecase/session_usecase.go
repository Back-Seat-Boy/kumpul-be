package usecase

import (
	"context"
	"time"

	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
)

type sessionUsecase struct {
	sessionRepo model.SessionRepository
}

func NewSessionUsecase(sessionRepo model.SessionRepository) model.SessionUsecase {
	return &sessionUsecase{sessionRepo: sessionRepo}
}

func (u *sessionUsecase) Create(ctx context.Context, user *model.User) (*model.Session, error) {
	session := &model.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		ExpiresAt: time.Now().Add(config.CacheTTL()),
	}

	if err := u.sessionRepo.Create(ctx, session, config.CacheTTL()); err != nil {
		log.WithFields(log.Fields{
			"context": utils.DumpIncomingContext(ctx),
			"session": utils.Dump(session),
		}).Error(err)

		return nil, err
	}

	return session, nil
}

func (u *sessionUsecase) GetByID(ctx context.Context, sessionID string) (*model.Session, error) {
	return u.sessionRepo.FindByID(ctx, sessionID)
}

func (u *sessionUsecase) Delete(ctx context.Context, sessionID string) error {
	return u.sessionRepo.Delete(ctx, sessionID)
}
