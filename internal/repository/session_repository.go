package repository

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/cacher"
	"github.com/kumparan/go-utils"
)

type sessionRepo struct {
	keeper cacher.Keeper
}

func NewSessionRepository(keeper cacher.Keeper) model.SessionRepository {
	return &sessionRepo{keeper: keeper}
}

func (r *sessionRepo) Create(ctx context.Context, session *model.Session, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", session.ID)
	if err := r.keeper.StoreWithoutBlocking(cacher.NewItemWithCustomTTL(key, utils.ToByte(session), ttl)); err != nil {
		log.WithFields(log.Fields{
			"context": utils.DumpIncomingContext(ctx),
			"session": utils.Dump(session),
		}).Error(err)
		return err
	}
	return nil
}

func (r *sessionRepo) FindByID(ctx context.Context, sessionID string) (*model.Session, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"sessionID": sessionID,
	})

	key := fmt.Sprintf("session:%s", sessionID)
	session, mu, err := findFromCacheByKey[*model.Session](r.keeper, key)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	defer cacher.SafeUnlock(mu)

	if session == nil {
		return nil, model.ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		err = r.Delete(ctx, sessionID)
		if err != nil {
			logger.Error(err)
		}
		return nil, model.ErrSessionExpired
	}

	return session, nil
}

func (r *sessionRepo) Delete(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := r.keeper.DeleteByKeys([]string{key}); err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"sessionID": sessionID,
			"key":       key,
		})
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
