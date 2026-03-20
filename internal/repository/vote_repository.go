package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type voteRepo struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) model.VoteRepository {
	return &voteRepo{db: db}
}

func (r *voteRepo) FindByEventOptionIDAndUserID(ctx context.Context, eventOptionID, userID string) (*model.Vote, error) {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"eventOptionID": eventOptionID,
		"userID":        userID,
	})

	var vote model.Vote
	if err := r.db.WithContext(ctx).Where("event_option_id = ? AND user_id = ?", eventOptionID, userID).First(&vote).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrVoteNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find vote: %w", err)
	}
	return &vote, nil
}

func (r *voteRepo) Create(ctx context.Context, vote *model.Vote) error {
	if err := r.db.WithContext(ctx).Create(vote).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":  utils.DumpIncomingContext(ctx),
			"vote": utils.Dump(vote),
		}).Error(err)
		return fmt.Errorf("failed to create vote: %w", err)
	}
	return nil
}

func (r *voteRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Vote{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete vote: %w", err)
	}
	return nil
}
