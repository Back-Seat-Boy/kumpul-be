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
			return nil, model.ErrEventNotFound
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
			return nil, model.ErrEventNotFound
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

func (r *eventRepo) FindByParticipantUserID(ctx context.Context, userID string) ([]*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"userID":  userID,
	})

	var events []*model.Event
	if err := r.db.WithContext(ctx).
		Model(&model.Event{}).
		Preload("Creator").
		Joins("JOIN participants ON participants.event_id = events.id").
		Where("participants.user_id = ?", userID).
		Group("events.id").
		Order("events.created_at desc").
		Find(&events).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list participated events: %w", err)
	}
	return events, nil
}

func (r *eventRepo) List(ctx context.Context) ([]*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
	})

	var events []*model.Event
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&events).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	return events, nil
}

func (r *eventRepo) ListPaginated(ctx context.Context, req *model.ListEventsRequest) ([]*model.Event, int64, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"req":     req,
	})

	query := r.db.WithContext(ctx).Model(&model.Event{}).Preload("Creator")
	if req.RequesterUserID != "" {
		query = query.Where(
			`(
				events.visibility = ?
				OR events.created_by = ?
				OR EXISTS (
					SELECT 1
					FROM participants
					WHERE participants.event_id = events.id
						AND participants.user_id = ?
				)
				OR EXISTS (
					SELECT 1
					FROM event_options eo
					JOIN votes v ON v.event_option_id = eo.id
					WHERE eo.event_id = events.id
						AND v.user_id = ?
				)
			)`,
			model.EventVisibilityPublic,
			req.RequesterUserID,
			req.RequesterUserID,
			req.RequesterUserID,
		)
	}
	if req.Filter.Search != "" {
		query = query.Where("title ILIKE ?", "%"+req.Filter.Search+"%")
	}

	if req.Filter.Status != "" {
		query = query.Where("status = ?", req.Filter.Status)
	}
	if req.Filter.Visibility != "" {
		query = query.Where("visibility = ?", req.Filter.Visibility)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	query = query.Order("created_at DESC")
	if req.Mode == model.PaginationModeCursor && req.Cursor != "" {
		var cursorEvent model.Event
		if err := r.db.WithContext(ctx).First(&cursorEvent, "id = ?", req.Cursor).Error; err != nil {
			logger.Error(err)
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		query = query.Where("created_at <= ? AND id != ?", cursorEvent.CreatedAt, req.Cursor)
	} else {
		if req.Page <= 0 {
			req.Page = 1
		}
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset)
	}

	query = query.Limit(req.Limit)

	var events []*model.Event
	if err := query.Find(&events).Error; err != nil {
		logger.Error(err)
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}

	return events, total, nil
}

func (r *eventRepo) BulkFetchVoteCounts(ctx context.Context, eventIDs []string) (map[string]int64, error) {
	if len(eventIDs) == 0 {
		return map[string]int64{}, nil
	}

	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"eventIDs": eventIDs,
	})

	type result struct {
		EventID   string `gorm:"column:event_id"`
		VoteCount int64  `gorm:"column:vote_count"`
	}

	var results []result
	query := `
		SELECT eo.event_id, COUNT(v.id) as vote_count
		FROM event_options eo
		LEFT JOIN votes v ON eo.id = v.event_option_id
		WHERE eo.event_id IN ?
		GROUP BY eo.event_id
	`

	if err := r.db.WithContext(ctx).Raw(query, eventIDs).Scan(&results).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to fetch vote counts: %w", err)
	}

	counts := make(map[string]int64, len(results))
	for _, r := range results {
		counts[r.EventID] = r.VoteCount
	}
	return counts, nil
}

func (r *eventRepo) BulkFetchParticipantCounts(ctx context.Context, eventIDs []string) (map[string]int64, error) {
	if len(eventIDs) == 0 {
		return map[string]int64{}, nil
	}

	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"eventIDs": eventIDs,
	})

	type result struct {
		EventID          string `gorm:"column:event_id"`
		ParticipantCount int64  `gorm:"column:participant_count"`
	}

	var results []result
	query := `
		SELECT event_id, COUNT(*) as participant_count
		FROM participants
		WHERE event_id IN ?
		GROUP BY event_id
	`

	if err := r.db.WithContext(ctx).Raw(query, eventIDs).Scan(&results).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to fetch participant counts: %w", err)
	}

	counts := make(map[string]int64, len(results))
	for _, r := range results {
		counts[r.EventID] = r.ParticipantCount
	}
	return counts, nil
}

func (r *eventRepo) BulkFetchChosenOptionDetails(ctx context.Context, chosenOptionIDs []string) (map[string]*model.ChosenOptionDetails, error) {
	if len(chosenOptionIDs) == 0 {
		return map[string]*model.ChosenOptionDetails{}, nil
	}

	logger := log.WithFields(log.Fields{
		"context":         utils.DumpIncomingContext(ctx),
		"chosenOptionIDs": chosenOptionIDs,
	})

	type result struct {
		OptionID  string `gorm:"column:option_id"`
		Date      string `gorm:"column:date"`
		StartTime string `gorm:"column:start_time"`
		EndTime   string `gorm:"column:end_time"`
		VenueName string `gorm:"column:venue_name"`
	}

	var results []result
	query := `
		SELECT eo.id as option_id, eo.date::text, eo.start_time, eo.end_time, v.name as venue_name
		FROM event_options eo
		JOIN venues v ON eo.venue_id = v.id
		WHERE eo.id IN ?
	`

	if err := r.db.WithContext(ctx).Raw(query, chosenOptionIDs).Scan(&results).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to fetch option details: %w", err)
	}

	details := make(map[string]*model.ChosenOptionDetails, len(results))
	for _, r := range results {
		details[r.OptionID] = &model.ChosenOptionDetails{
			Date:      r.Date,
			StartTime: r.StartTime,
			EndTime:   r.EndTime,
			VenueName: r.VenueName,
		}
	}
	return details, nil
}

func (r *eventRepo) BulkFetchPaymentRecordCounts(ctx context.Context, eventIDs []string) (map[string]*model.PaymentRecordCounts, error) {
	if len(eventIDs) == 0 {
		return map[string]*model.PaymentRecordCounts{}, nil
	}

	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"eventIDs": eventIDs,
	})

	type result struct {
		EventID string `gorm:"column:event_id"`
		Status  string `gorm:"column:status"`
		Count   int64  `gorm:"column:count"`
	}

	var results []result
	query := `
		SELECT p.event_id, pr.status, COUNT(*) as count
		FROM payments p
		JOIN payment_records pr ON p.id = pr.payment_id
		WHERE p.event_id IN ?
		GROUP BY p.event_id, pr.status
	`

	if err := r.db.WithContext(ctx).Raw(query, eventIDs).Scan(&results).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to fetch payment counts: %w", err)
	}

	counts := make(map[string]*model.PaymentRecordCounts)
	for _, r := range results {
		if _, ok := counts[r.EventID]; !ok {
			counts[r.EventID] = &model.PaymentRecordCounts{}
		}
		switch model.PaymentRecordStatus(r.Status) {
		case model.PaymentRecordStatusPending:
			counts[r.EventID].Pending = r.Count
		case model.PaymentRecordStatusClaimed:
			counts[r.EventID].Claimed = r.Count
		case model.PaymentRecordStatusConfirmed:
			counts[r.EventID].Confirmed = r.Count
		}
	}
	return counts, nil
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

func (r *eventRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, event *model.Event) error {
	if err := tx.WithContext(ctx).Create(event).Error; err != nil {
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

func (r *eventRepo) UpdateStatusWithTx(ctx context.Context, tx *gorm.DB, id string, status model.EventStatus) error {
	if err := tx.WithContext(ctx).Model(&model.Event{}).Where("id = ?", id).Update("status", status).Error; err != nil {
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

func (r *eventRepo) UpdateChosenOptionWithTx(ctx context.Context, tx *gorm.DB, id string, optionID string) error {
	if err := tx.WithContext(ctx).Model(&model.Event{}).Where("id = ?", id).Update("chosen_option_id", optionID).Error; err != nil {
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
