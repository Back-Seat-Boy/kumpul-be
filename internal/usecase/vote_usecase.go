package usecase

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type voteUsecase struct {
	voteRepo       model.VoteRepository
	eventRepo      model.EventRepository
	eventOptionRepo model.EventOptionRepository
}

func NewVoteUsecase(voteRepo model.VoteRepository, eventRepo model.EventRepository, eventOptionRepo model.EventOptionRepository) model.VoteUsecase {
	return &voteUsecase{voteRepo: voteRepo, eventRepo: eventRepo, eventOptionRepo: eventOptionRepo}
}

func (u *voteUsecase) CastVote(ctx context.Context, userID string, req *model.CastVoteRequest) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"eventOptionID": req.EventOptionID,
		"userID":        userID,
	})

	// Get event option to find event ID
	eventOption, err := u.eventOptionRepo.FindByID(ctx, req.EventOptionID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Check event status - can only vote when status is "voting"
	event, err := u.eventRepo.FindByID(ctx, eventOption.EventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if event.Status != model.EventStatusVoting {
		return model.ErrEventNotInVotingPhase
	}

	_, err = u.voteRepo.FindByEventOptionIDAndUserID(ctx, req.EventOptionID, userID)
	if err == nil {
		return model.ErrAlreadyVoted
	}

	vote := &model.Vote{
		ID:            uuid.New().String(),
		EventOptionID: req.EventOptionID,
		UserID:        userID,
	}

	if err := u.voteRepo.Create(ctx, vote); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *voteUsecase) RemoveVote(ctx context.Context, eventOptionID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"eventOptionID": eventOptionID,
		"userID":        userID,
	})

	// Get event option to find event ID
	eventOption, err := u.eventOptionRepo.FindByID(ctx, eventOptionID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Check event status - can only remove vote when status is "voting"
	event, err := u.eventRepo.FindByID(ctx, eventOption.EventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if event.Status != model.EventStatusVoting {
		return model.ErrEventNotInVotingPhase
	}

	vote, err := u.voteRepo.FindByEventOptionIDAndUserID(ctx, eventOptionID, userID)
	if err != nil {
		if err == model.ErrVoteNotFound {
			return model.ErrNotVoted
		}
		logger.Error(err)
		return err
	}

	if err := u.voteRepo.Delete(ctx, vote.ID); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
