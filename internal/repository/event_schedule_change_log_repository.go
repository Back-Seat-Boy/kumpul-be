package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type eventScheduleChangeLogRepo struct {
	db *gorm.DB
}

func NewEventScheduleChangeLogRepository(db *gorm.DB) model.EventScheduleChangeLogRepository {
	return &eventScheduleChangeLogRepo{db: db}
}

func (r *eventScheduleChangeLogRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, changeLog *model.EventScheduleChangeLog) error {
	if err := tx.WithContext(ctx).Create(changeLog).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"log":    utils.Dump(changeLog),
			"source": "eventScheduleChangeLogRepo.CreateWithTx",
		}).Error(err)
		return fmt.Errorf("failed to create event schedule change log: %w", err)
	}
	return nil
}

func (r *eventScheduleChangeLogRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.EventScheduleChangeLog, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var logs []*model.EventScheduleChangeLog
	if err := r.db.WithContext(ctx).
		Preload("Editor").
		Preload("OldVenue").
		Preload("NewVenue").
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list event schedule change logs: %w", err)
	}

	return logs, nil
}
