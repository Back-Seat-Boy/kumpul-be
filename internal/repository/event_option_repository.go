package repository

import (
	"context"
	"fmt"
	"time"

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

type eventOptionWithVoteCountResult struct {
	model.EventOption
	VoteCount int64 `json:"vote_count"`
	HasVoted  bool  `json:"has_voted"`
}

func (r *eventOptionRepo) FindByEventIDWithVoteCount(ctx context.Context, eventID string, userID *string) ([]*model.EventOptionWithVoteCount, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	var rawResults []*eventOptionWithVoteCountResult
	var query string
	var args []interface{}

	if userID != nil && *userID != "" {
		query = `
			SELECT eo.*, 
				COUNT(v.id) as vote_count,
				CASE WHEN MAX(CASE WHEN v.user_id = ? THEN 1 ELSE 0 END) = 1 THEN true ELSE false END as has_voted
			FROM event_options eo
			LEFT JOIN votes v ON eo.id = v.event_option_id
			WHERE eo.event_id = ?
			GROUP BY eo.id
		`
		args = []interface{}{*userID, eventID}
	} else {
		query = `
			SELECT eo.*, COUNT(v.id) as vote_count, false as has_voted
			FROM event_options eo
			LEFT JOIN votes v ON eo.id = v.event_option_id
			WHERE eo.event_id = ?
			GROUP BY eo.id
		`
		args = []interface{}{eventID}
	}

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rawResults).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list event options with vote count: %w", err)
	}

	// Convert to the public type and preload venues
	results := make([]*model.EventOptionWithVoteCount, len(rawResults))
	for i, raw := range rawResults {
		results[i] = &model.EventOptionWithVoteCount{
			EventOption: raw.EventOption,
			VoteCount:   raw.VoteCount,
			HasVoted:    raw.HasVoted,
		}
	}

	// Preload venues for all results
	for _, result := range results {
		var venue model.Venue
		if err := r.db.WithContext(ctx).First(&venue, "id = ?", result.VenueID).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				logger.Error(err)
			}
		} else {
			result.Venue = venue
		}
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

func (r *eventOptionRepo) BulkCreateWithTx(ctx context.Context, tx *gorm.DB, options []*model.EventOption) error {
	if err := tx.WithContext(ctx).Create(options).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"option": utils.Dump(options),
		}).Error(err)
		return fmt.Errorf("failed to create event options: %w", err)
	}
	return nil
}

func (r *eventOptionRepo) UpdateScheduleWithTx(ctx context.Context, tx *gorm.DB, optionID string, venueID string, date time.Time, startTime string, endTime string) error {
	if err := tx.WithContext(ctx).
		Model(&model.EventOption{}).
		Where("id = ?", optionID).
		Updates(map[string]interface{}{
			"venue_id":   venueID,
			"date":       date,
			"start_time": startTime,
			"end_time":   endTime,
		}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":      utils.DumpIncomingContext(ctx),
			"optionID": optionID,
		}).Error(err)
		return fmt.Errorf("failed to update event option schedule: %w", err)
	}
	return nil
}

func (r *eventOptionRepo) UpdateWithTx(ctx context.Context, tx *gorm.DB, optionID string, venueID string, date time.Time, startTime string, endTime string) error {
	if err := tx.WithContext(ctx).
		Model(&model.EventOption{}).
		Where("id = ?", optionID).
		Updates(map[string]interface{}{
			"venue_id":   venueID,
			"date":       date,
			"start_time": startTime,
			"end_time":   endTime,
		}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":      utils.DumpIncomingContext(ctx),
			"optionID": optionID,
		}).Error(err)
		return fmt.Errorf("failed to update event option: %w", err)
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

func (r *eventOptionRepo) FindVotersByOptionID(ctx context.Context, optionID string) ([]model.VoterInfo, error) {
	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"optionID": optionID,
	})

	var voters []model.VoterInfo
	query := `
		SELECT v.user_id, u.name as user_name, u.avatar_url
		FROM votes v
		JOIN users u ON v.user_id = u.id
		WHERE v.event_option_id = ?
		ORDER BY v.created_at ASC
	`
	if err := r.db.WithContext(ctx).Raw(query, optionID).Scan(&voters).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to find voters: %w", err)
	}
	return voters, nil
}

func (r *eventOptionRepo) FindVotersByOptionIDs(ctx context.Context, optionIDs []string) ([]model.VoterInfo, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"optionIDs": optionIDs,
	})

	var voters []model.VoterInfo
	query := `
		SELECT v.event_option_id, v.user_id, u.name as user_name, u.avatar_url
		FROM votes v
		JOIN users u ON v.user_id = u.id
		WHERE v.event_option_id IN (?)
		ORDER BY v.created_at ASC
	`
	if err := r.db.WithContext(ctx).Raw(query, optionIDs).Scan(&voters).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to find voters: %w", err)
	}
	return voters, nil
}
