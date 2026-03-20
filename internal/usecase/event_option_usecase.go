package usecase

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
)

type eventOptionUsecase struct {
	optionRepo model.EventOptionRepository
}

func NewEventOptionUsecase(optionRepo model.EventOptionRepository) model.EventOptionUsecase {
	return &eventOptionUsecase{optionRepo: optionRepo}
}

func (u *eventOptionUsecase) GetByID(ctx context.Context, id string) (*model.EventOption, error) {
	return u.optionRepo.FindByID(ctx, id)
}

func (u *eventOptionUsecase) ListByEvent(ctx context.Context, eventID string) ([]*model.EventOptionWithVoteCount, error) {
	return u.optionRepo.FindByEventIDWithVoteCount(ctx, eventID)
}

func (u *eventOptionUsecase) Create(ctx context.Context, eventID string, req *model.CreateEventOptionRequest) (*model.EventOption, error) {
	option := &model.EventOption{
		ID:        uuid.New().String(),
		EventID:   eventID,
		VenueID:   req.VenueID,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := u.optionRepo.Create(ctx, option); err != nil {
		return nil, err
	}

	return option, nil
}

func (u *eventOptionUsecase) Delete(ctx context.Context, id string) error {
	return u.optionRepo.Delete(ctx, id)
}
