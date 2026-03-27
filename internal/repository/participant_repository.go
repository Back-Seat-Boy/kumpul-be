package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type participantRepo struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) model.ParticipantRepository {
	return &participantRepo{db: db}
}

func (r *participantRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var participants []*model.Participant
	if err := r.db.WithContext(ctx).Preload("User").Where("event_id = ?", eventID).Find(&participants).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}
	return participants, nil
}

func (r *participantRepo) FindByEventIDAndUserID(ctx context.Context, eventID, userID string) (*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	var participant model.Participant
	if err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrParticipantNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find participant: %w", err)
	}
	return &participant, nil
}

func (r *participantRepo) Create(ctx context.Context, participant *model.Participant) error {
	if err := r.db.WithContext(ctx).Create(participant).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"participant": utils.Dump(participant),
		}).Error(err)
		return fmt.Errorf("failed to create participant: %w", err)
	}
	return nil
}

func (r *participantRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, participant *model.Participant) error {
	if err := tx.WithContext(ctx).Create(participant).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"participant": utils.Dump(participant),
		}).Error(err)
		return fmt.Errorf("failed to create participant: %w", err)
	}
	return nil
}

func (r *participantRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Participant{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete participant: %w", err)
	}
	return nil
}

func (r *participantRepo) CountByEventID(ctx context.Context, eventID string) (int64, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Participant{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
		logger.Error(err)
		return 0, fmt.Errorf("failed to count participants: %w", err)
	}
	return count, nil
}
