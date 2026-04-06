package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type eventOptionChangeLogRepo struct {
	db *gorm.DB
}

func NewEventOptionChangeLogRepository(db *gorm.DB) model.EventOptionChangeLogRepository {
	return &eventOptionChangeLogRepo{db: db}
}

func (r *eventOptionChangeLogRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, changeLog *model.EventOptionChangeLog) error {
	if err := tx.WithContext(ctx).Create(changeLog).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"log":    utils.Dump(changeLog),
			"source": "eventOptionChangeLogRepo.CreateWithTx",
		}).Error(err)
		return fmt.Errorf("failed to create event option change log: %w", err)
	}
	return nil
}

func (r *eventOptionChangeLogRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.EventOptionChangeLog, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var logs []*model.EventOptionChangeLog
	if err := r.db.WithContext(ctx).
		Preload("Editor").
		Preload("OldVenue").
		Preload("NewVenue").
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list event option change logs: %w", err)
	}

	return logs, nil
}
