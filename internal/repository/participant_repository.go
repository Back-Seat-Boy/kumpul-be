package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type participantRepo struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) model.ParticipantRepository {
	return &participantRepo{db: db}
}

const participantDisplayNameExpr = "COALESCE(NULLIF(users.name, ''), participants.guest_name)"

func (r *participantRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var participants []*model.Participant
	if err := r.db.WithContext(ctx).Preload("User").Where("event_id = ?", eventID).Find(&participants).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}
	for _, participant := range participants {
		participant.SetDerivedFields()
	}
	return participants, nil
}

func (r *participantRepo) ListPaginatedByEvent(ctx context.Context, req *model.ListParticipantsRequest) ([]*model.Participant, int64, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"req":     req,
	})

	query := r.db.WithContext(ctx).
		Model(&model.Participant{}).
		Preload("User").
		Joins("LEFT JOIN users ON users.id = participants.user_id").
		Where("participants.event_id = ?", req.EventID)

	if req.Filter.Search != "" {
		query = query.Where(participantDisplayNameExpr+" ILIKE ?", "%"+req.Filter.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, fmt.Errorf("failed to count participants: %w", err)
	}

	sortOrder := strings.ToUpper(string(req.SortOrder))
	if sortOrder != "DESC" {
		sortOrder = "ASC"
	}

	query = query.Order(participantDisplayNameExpr + " " + sortOrder).Order("participants.id " + sortOrder)

	if req.Mode == model.PaginationModeCursor && req.Cursor != "" {
		type participantCursor struct {
			ID          string
			DisplayName string
		}

		var cursor participantCursor
		cursorQuery := r.db.WithContext(ctx).
			Model(&model.Participant{}).
			Select("participants.id", participantDisplayNameExpr+" AS display_name").
			Joins("LEFT JOIN users ON users.id = participants.user_id").
			Where("participants.event_id = ? AND participants.id = ?", req.EventID, req.Cursor)

		if err := cursorQuery.First(&cursor).Error; err != nil {
			logger.Error(err)
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}

		comparator := ">"
		if sortOrder == "DESC" {
			comparator = "<"
		}

		query = query.Where(
			fmt.Sprintf("(%s %s ? OR (%s = ? AND participants.id %s ?))", participantDisplayNameExpr, comparator, participantDisplayNameExpr, comparator),
			cursor.DisplayName,
			cursor.DisplayName,
			cursor.ID,
		)
	} else {
		if req.Page <= 0 {
			req.Page = 1
		}
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset)
	}

	query = query.Limit(req.Limit)

	var participants []*model.Participant
	if err := query.Find(&participants).Error; err != nil {
		logger.Error(err)
		return nil, 0, fmt.Errorf("failed to list participants: %w", err)
	}

	for _, participant := range participants {
		participant.SetDerivedFields()
	}

	return participants, total, nil
}

func (r *participantRepo) FindByEventIDAndUserID(ctx context.Context, eventID, userID string) (*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	var participant model.Participant
	if err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrParticipantNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find participant: %w", err)
	}
	participant.SetDerivedFields()
	return &participant, nil
}

func (r *participantRepo) Create(ctx context.Context, participant *model.Participant) error {
	if err := r.db.WithContext(ctx).Create(participant).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"participant": utils.Dump(participant),
		}).Error(err)
		return fmt.Errorf("failed to create participant: %w", err)
	}
	return nil
}

func (r *participantRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, participant *model.Participant) error {
	if err := tx.WithContext(ctx).Create(participant).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"participant": utils.Dump(participant),
		}).Error(err)
		return fmt.Errorf("failed to create participant: %w", err)
	}
	return nil
}

func (r *participantRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Participant{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete participant: %w", err)
	}
	return nil
}

func (r *participantRepo) CountByEventID(ctx context.Context, eventID string) (int64, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Participant{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
		logger.Error(err)
		return 0, fmt.Errorf("failed to count participants: %w", err)
	}
	return count, nil
}

func (r *participantRepo) FindByEventIDAndGuestName(ctx context.Context, eventID, guestName string) (*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"eventID":   eventID,
		"guestName": guestName,
	})

	var participant model.Participant
	if err := r.db.WithContext(ctx).Where("event_id = ? AND guest_name = ?", eventID, guestName).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrParticipantNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find participant: %w", err)
	}
	participant.SetDerivedFields()
	return &participant, nil
}

func (r *participantRepo) FindByID(ctx context.Context, id string) (*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})

	var participant model.Participant
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrParticipantNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find participant: %w", err)
	}
	participant.SetDerivedFields()
	return &participant, nil
}

func (r *participantRepo) FindByEventIDAndID(ctx context.Context, eventID, id string) (*model.Participant, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"id":      id,
	})

	var participant model.Participant
	if err := r.db.WithContext(ctx).Where("event_id = ? AND id = ?", eventID, id).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrParticipantNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find participant: %w", err)
	}
	participant.SetDerivedFields()
	return &participant, nil
}
