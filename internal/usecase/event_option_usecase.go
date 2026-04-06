package usecase

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
)

type eventOptionUsecase struct {
	optionRepo        model.EventOptionRepository
	eventRepo         model.EventRepository
	venueRepo         model.VenueRepository
	changeLogRepo     model.EventOptionChangeLogRepository
	gormTransactioner model.GormTransactioner
}

func NewEventOptionUsecase(optionRepo model.EventOptionRepository, eventRepo model.EventRepository, venueRepo model.VenueRepository, changeLogRepo model.EventOptionChangeLogRepository, gormTransactioner model.GormTransactioner) model.EventOptionUsecase {
	return &eventOptionUsecase{
		optionRepo:        optionRepo,
		eventRepo:         eventRepo,
		venueRepo:         venueRepo,
		changeLogRepo:     changeLogRepo,
		gormTransactioner: gormTransactioner,
	}
}

func (u *eventOptionUsecase) GetByID(ctx context.Context, id string) (*model.EventOption, error) {
	return u.optionRepo.FindByID(ctx, id)
}

func (u *eventOptionUsecase) ListByEvent(ctx context.Context, eventID string, userID *string) ([]*model.EventOptionWithVoteCount, error) {
	return u.optionRepo.FindByEventIDWithVoteCount(ctx, eventID, userID)
}

func (u *eventOptionUsecase) ListByEventWithVoters(ctx context.Context, eventID string, userID *string) ([]*model.EventOptionWithVoteCount, error) {
	options, err := u.optionRepo.FindByEventIDWithVoteCount(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	var optionIDs []string
	optionsMap := make(map[string]*model.EventOptionWithVoteCount)
	for _, option := range options {
		optionIDs = append(optionIDs, option.ID)
		optionsMap[option.ID] = option
	}

	// Fetch voters for each option
	voters, err := u.optionRepo.FindVotersByOptionIDs(ctx, optionIDs)
	if err != nil {
		return nil, err
	}

	for i := range voters {
		if optionsMap[voters[i].EventOptionID] != nil {
			optionsMap[voters[i].EventOptionID].Voters = append(optionsMap[voters[i].EventOptionID].Voters, voters[i])
		}
	}

	return options, nil
}

func (u *eventOptionUsecase) ListChangeLogs(ctx context.Context, eventID string, userID string) ([]*model.EventOptionChangeLog, error) {
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return nil, err
	}
	if event.CreatedBy != userID {
		return nil, model.ErrForbidden
	}

	return u.changeLogRepo.FindByEventID(ctx, eventID)
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

func (u *eventOptionUsecase) Update(ctx context.Context, eventID string, optionID string, userID string, req *model.UpdateEventOptionRequest) error {
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}
	if event.CreatedBy != userID {
		return model.ErrForbidden
	}
	if event.Status != model.EventStatusVoting {
		return model.ErrEventOptionEditNotAllowed
	}

	option, err := u.optionRepo.FindByID(ctx, optionID)
	if err != nil {
		return err
	}
	if option.EventID != eventID {
		return model.ErrEventOptionNotFound
	}

	if _, err := u.venueRepo.FindByID(ctx, req.VenueID); err != nil {
		return err
	}

	tx := u.gormTransactioner.Begin(ctx)
	if err := u.optionRepo.UpdateWithTx(ctx, tx, optionID, req.VenueID, req.Date, req.StartTime, req.EndTime); err != nil {
		u.gormTransactioner.Rollback(tx)
		return err
	}

	changeLog := &model.EventOptionChangeLog{
		ID:            uuid.New().String(),
		EventID:       eventID,
		EventOptionID: optionID,
		EditedBy:      userID,
		Note:          req.Note,
		OldVenueID:    option.VenueID,
		OldDate:       option.Date,
		OldStartTime:  option.StartTime,
		OldEndTime:    option.EndTime,
		NewVenueID:    req.VenueID,
		NewDate:       req.Date,
		NewStartTime:  req.StartTime,
		NewEndTime:    req.EndTime,
	}
	if err := u.changeLogRepo.CreateWithTx(ctx, tx, changeLog); err != nil {
		u.gormTransactioner.Rollback(tx)
		return err
	}

	return u.gormTransactioner.Commit(tx)
}

func (u *eventOptionUsecase) Delete(ctx context.Context, id string) error {
	return u.optionRepo.Delete(ctx, id)
}
