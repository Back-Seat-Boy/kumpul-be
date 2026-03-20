package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type venueRepo struct {
	db *gorm.DB
}

func NewVenueRepository(db *gorm.DB) model.VenueRepository {
	return &venueRepo{db: db}
}

func (r *venueRepo) FindByID(ctx context.Context, id string) (*model.Venue, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})

	var venue model.Venue
	if err := r.db.WithContext(ctx).Preload("Creator").First(&venue, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrVenueNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find venue: %w", err)
	}
	return &venue, nil
}

func (r *venueRepo) FindByCreatedBy(ctx context.Context, createdBy string) ([]*model.Venue, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"createdBy": createdBy,
	})

	var venues []*model.Venue
	if err := r.db.WithContext(ctx).Where("created_by = ?", createdBy).Find(&venues).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list venues: %w", err)
	}
	return venues, nil
}

func (r *venueRepo) Create(ctx context.Context, venue *model.Venue) error {
	if err := r.db.WithContext(ctx).Create(venue).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"venue": utils.Dump(venue),
		}).Error(err)
		return fmt.Errorf("failed to create venue: %w", err)
	}
	return nil
}

func (r *venueRepo) Update(ctx context.Context, venue *model.Venue) error {
	if err := r.db.WithContext(ctx).Save(venue).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"venue": utils.Dump(venue),
		}).Error(err)
		return fmt.Errorf("failed to update venue: %w", err)
	}
	return nil
}

func (r *venueRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Venue{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete venue: %w", err)
	}
	return nil
}
