package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type participantGroupRepo struct {
	db *gorm.DB
}

func NewParticipantGroupRepository(db *gorm.DB) model.ParticipantGroupRepository {
	return &participantGroupRepo{db: db}
}

func (r *participantGroupRepo) ListByEventID(ctx context.Context, eventID string) ([]*model.ParticipantGroup, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var groups []*model.ParticipantGroup
	if err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&groups).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list participant groups: %w", err)
	}
	return groups, nil
}

func (r *participantGroupRepo) FindByID(ctx context.Context, id string) (*model.ParticipantGroup, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})

	var group model.ParticipantGroup
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrParticipantGroupNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find participant group: %w", err)
	}
	return &group, nil
}

func (r *participantGroupRepo) Create(ctx context.Context, group *model.ParticipantGroup) error {
	if err := r.db.WithContext(ctx).Create(group).Error; err != nil {
		if isUniqueConstraintViolation(err) {
			return model.ErrParticipantGroupNameTaken
		}
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"group": utils.Dump(group),
		}).Error(err)
		return fmt.Errorf("failed to create participant group: %w", err)
	}
	return nil
}

func (r *participantGroupRepo) Update(ctx context.Context, group *model.ParticipantGroup) error {
	if err := r.db.WithContext(ctx).
		Model(group).
		Select("name").
		Updates(group).Error; err != nil {
		if isUniqueConstraintViolation(err) {
			return model.ErrParticipantGroupNameTaken
		}
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"group": utils.Dump(group),
		}).Error(err)
		return fmt.Errorf("failed to update participant group: %w", err)
	}
	return nil
}

func (r *participantGroupRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.ParticipantGroup{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete participant group: %w", err)
	}
	return nil
}

func isUniqueConstraintViolation(err error) bool {
	var sqlStateErr interface{ SQLState() string }
	if errors.As(err, &sqlStateErr) {
		return sqlStateErr.SQLState() == "23505"
	}
	return strings.Contains(err.Error(), "23505")
}
