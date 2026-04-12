package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type eventImageRepo struct {
	db *gorm.DB
}

func NewEventImageRepository(db *gorm.DB) model.EventImageRepository {
	return &eventImageRepo{db: db}
}

func (r *eventImageRepo) ReplaceByEventIDWithTx(ctx context.Context, tx *gorm.DB, eventID string, images []*model.EventImage) error {
	if err := tx.WithContext(ctx).Where("event_id = ?", eventID).Delete(&model.EventImage{}).Error; err != nil {
		return fmt.Errorf("failed to delete event images: %w", err)
	}
	if len(images) == 0 {
		return nil
	}
	if err := tx.WithContext(ctx).Create(images).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":     utils.DumpIncomingContext(ctx),
			"eventID": eventID,
			"images":  utils.Dump(images),
		}).Error(err)
		return fmt.Errorf("failed to create event images: %w", err)
	}
	return nil
}

func (r *eventImageRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.EventImage, error) {
	var images []*model.EventImage
	if err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("position ASC").
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("failed to find event images: %w", err)
	}
	return images, nil
}
