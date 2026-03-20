package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type eventRepo struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) model.EventRepository {
	return &eventRepo{db: db}
}

func (r *eventRepo) FindByID(ctx context.Context, id string) (*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})

	var event model.Event
	if err := r.db.WithContext(ctx).Preload("Creator").Preload("ChosenOption").First(&event, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrUserNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find event: %w", err)
	}
	return &event, nil
}

func (r *eventRepo) FindByShareToken(ctx context.Context, token string) (*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"token":   token,
	})

	var event model.Event
	if err := r.db.WithContext(ctx).Preload("Creator").Preload("ChosenOption").First(&event, "share_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrUserNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find event: %w", err)
	}
	return &event, nil
}

func (r *eventRepo) FindByCreatedBy(ctx context.Context, createdBy string) ([]*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"createdBy": createdBy,
	})

	var events []*model.Event
	if err := r.db.WithContext(ctx).Where("created_by = ?", createdBy).Order("created_at desc").Find(&events).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	return events, nil
}

func (r *eventRepo) Create(ctx context.Context, event *model.Event) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"event": utils.Dump(event),
		}).Error(err)
		return fmt.Errorf("failed to create event: %w", err)
	}
	return nil
}

func (r *eventRepo) Update(ctx context.Context, event *model.Event) error {
	if err := r.db.WithContext(ctx).Save(event).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"event": utils.Dump(event),
		}).Error(err)
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

func (r *eventRepo) UpdateStatus(ctx context.Context, id string, status model.EventStatus) error {
	if err := r.db.WithContext(ctx).Model(&model.Event{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"id":     id,
			"status": status,
		}).Error(err)
		return fmt.Errorf("failed to update event status: %w", err)
	}
	return nil
}

func (r *eventRepo) UpdateChosenOption(ctx context.Context, id string, optionID string) error {
	if err := r.db.WithContext(ctx).Model(&model.Event{}).Where("id = ?", id).Update("chosen_option_id", optionID).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":      utils.DumpIncomingContext(ctx),
			"id":       id,
			"optionID": optionID,
		}).Error(err)
		return fmt.Errorf("failed to update chosen option: %w", err)
	}
	return nil
}

func (r *eventRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Event{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}
