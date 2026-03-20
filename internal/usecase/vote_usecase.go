package usecase

import (
	"context"
	"errors"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type voteUsecase struct {
	voteRepo model.VoteRepository
}

func NewVoteUsecase(voteRepo model.VoteRepository) model.VoteUsecase {
	return &voteUsecase{voteRepo: voteRepo}
}

func (u *voteUsecase) CastVote(ctx context.Context, userID string, req *model.CastVoteRequest) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"eventOptionID": req.EventOptionID,
		"userID":        userID,
	})

	_, err := u.voteRepo.FindByEventOptionIDAndUserID(ctx, req.EventOptionID, userID)
	if err == nil {
		return errors.New("already voted")
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

	vote, err := u.voteRepo.FindByEventOptionIDAndUserID(ctx, eventOptionID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := u.voteRepo.Delete(ctx, vote.ID); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
