package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type eventOptionRepo struct {
	db *gorm.DB
}

func NewEventOptionRepository(db *gorm.DB) model.EventOptionRepository {
	return &eventOptionRepo{db: db}
}

func (r *eventOptionRepo) FindByID(ctx context.Context, id string) (*model.EventOption, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})

	var option model.EventOption
	if err := r.db.WithContext(ctx).Preload("Event").Preload("Venue").First(&option, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrEventOptionNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find event option: %w", err)
	}
	return &option, nil
}

func (r *eventOptionRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.EventOption, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var options []*model.EventOption
	if err := r.db.WithContext(ctx).Preload("Venue").Where("event_id = ?", eventID).Find(&options).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list event options: %w", err)
	}
	return options, nil
}

func (r *eventOptionRepo) FindByEventIDWithVoteCount(ctx context.Context, eventID string) ([]*model.EventOptionWithVoteCount, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var results []*model.EventOptionWithVoteCount
	query := `
		SELECT eo.*, COUNT(v.id) as vote_count
		FROM event_options eo
		LEFT JOIN votes v ON eo.id = v.event_option_id
		WHERE eo.event_id = ?
		GROUP BY eo.id
	`
	if err := r.db.WithContext(ctx).Raw(query, eventID).Scan(&results).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list event options with vote count: %w", err)
	}
	return results, nil
}

func (r *eventOptionRepo) Create(ctx context.Context, option *model.EventOption) error {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"option": utils.Dump(option),
		}).Error(err)
		return fmt.Errorf("failed to create event option: %w", err)
	}
	return nil
}

func (r *eventOptionRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.EventOption{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete event option: %w", err)
	}
	return nil
}
