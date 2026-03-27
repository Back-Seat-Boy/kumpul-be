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

func (r *venueRepo) ListAll(ctx context.Context) ([]*model.Venue, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
	})

	var venues []*model.Venue
	if err := r.db.WithContext(ctx).Preload("Creator").Order("name asc").Find(&venues).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list venues: %w", err)
	}
	return venues, nil
}

// ListPaginated returns paginated venues with filtering done at SQL level
func (r *venueRepo) ListPaginated(ctx context.Context, req *model.ListVenuesRequest) ([]*model.Venue, int64, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"req":     req,
	})

	// Build base query with filters
	query := r.db.WithContext(ctx).Model(&model.Venue{}).Preload("Creator")

	// Apply search filter (name or address ILIKE)
	if req.Filter.Search != "" {
		searchPattern := "%" + req.Filter.Search + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, fmt.Errorf("failed to count venues: %w", err)
	}

	// Apply pagination
	query = query.Order("name ASC")

	if req.Mode == model.PaginationModeCursor && req.Cursor != "" {
		// Cursor-based: get venues with name >= cursor's venue name
		var cursorVenue model.Venue
		if err := r.db.WithContext(ctx).First(&cursorVenue, "id = ?", req.Cursor).Error; err != nil {
			logger.Error(err)
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		query = query.Where("(name > ? OR (name = ? AND id > ?))", cursorVenue.Name, cursorVenue.Name, req.Cursor)
	} else {
		// Page-based offset
		if req.Page <= 0 {
			req.Page = 1
		}
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset)
	}

	// Apply limit
	query = query.Limit(req.Limit)

	// Execute query
	var venues []*model.Venue
	if err := query.Find(&venues).Error; err != nil {
		logger.Error(err)
		return nil, 0, fmt.Errorf("failed to list venues: %w", err)
	}

	return venues, total, nil
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
