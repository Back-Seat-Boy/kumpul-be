package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type eventUsecase struct {
	eventRepo model.EventRepository
}

func NewEventUsecase(eventRepo model.EventRepository) model.EventUsecase {
	return &eventUsecase{eventRepo: eventRepo}
}

func (u *eventUsecase) GetByID(ctx context.Context, id string) (*model.Event, error) {
	return u.eventRepo.FindByID(ctx, id)
}

func (u *eventUsecase) GetByShareToken(ctx context.Context, token string) (*model.Event, error) {
	return u.eventRepo.FindByShareToken(ctx, token)
}

func (u *eventUsecase) List(ctx context.Context) ([]*model.Event, error) {
	return u.eventRepo.List(ctx)
}

func (u *eventUsecase) Create(ctx context.Context, userID string, req *model.CreateEventRequest) (*model.Event, error) {
	event := &model.Event{
		ID:             uuid.New().String(),
		CreatedBy:      userID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         model.EventStatusVoting,
		ShareToken:     generateShareToken(),
		PlayerCap:      req.PlayerCap,
		VotingDeadline: req.VotingDeadline,
	}

	if err := u.eventRepo.Create(ctx, event); err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"event": utils.Dump(event),
		}).Error(err)
		return nil, err
	}

	return event, nil
}

func (u *eventUsecase) UpdateStatus(ctx context.Context, id string, status model.EventStatus) error {
	return u.eventRepo.UpdateStatus(ctx, id, status)
}

func (u *eventUsecase) UpdateChosenOption(ctx context.Context, id string, optionID string) error {
	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"id":       id,
		"optionID": optionID,
	})

	if err := u.eventRepo.UpdateChosenOption(ctx, id, optionID); err != nil {
		logger.Error(err)
		return err
	}

	return u.eventRepo.UpdateStatus(ctx, id, model.EventStatusConfirmed)
}

func (u *eventUsecase) Delete(ctx context.Context, id string) error {
	return u.eventRepo.Delete(ctx, id)
}

func generateShareToken() string {
	b := make([]byte, 15)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:10]
}
