package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type participantUsecase struct {
	participantRepo model.ParticipantRepository
}

func NewParticipantUsecase(participantRepo model.ParticipantRepository) model.ParticipantUsecase {
	return &participantUsecase{participantRepo: participantRepo}
}

func (u *participantUsecase) ListByEvent(ctx context.Context, eventID string) ([]*model.Participant, error) {
	return u.participantRepo.FindByEventID(ctx, eventID)
}

func (u *participantUsecase) Join(ctx context.Context, eventID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	_, err := u.participantRepo.FindByEventIDAndUserID(ctx, eventID, userID)
	if err == nil {
		return errors.New("already joined")
	}

	participant := &model.Participant{
		ID:       uuid.New().String(),
		EventID:  eventID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := u.participantRepo.Create(ctx, participant); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) Leave(ctx context.Context, eventID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	participant, err := u.participantRepo.FindByEventIDAndUserID(ctx, eventID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := u.participantRepo.Delete(ctx, participant.ID); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) GetParticipantCount(ctx context.Context, eventID string) (int64, error) {
	return u.participantRepo.CountByEventID(ctx, eventID)
}
