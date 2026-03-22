package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type eventUsecase struct {
	eventRepo         model.EventRepository
	gormTransactioner model.GormTransactioner
	eventOptionRepo   model.EventOptionRepository
	participantRepo   model.ParticipantRepository
	paymentRepo       model.PaymentRepository
	paymentRecordRepo model.PaymentRecordRepository
	venueRepo         model.VenueRepository
}

func NewEventUsecase(eventRepo model.EventRepository, gormTransactioner model.GormTransactioner, eventOptionRepo model.EventOptionRepository, participantRepo model.ParticipantRepository, paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, venueRepo model.VenueRepository) model.EventUsecase {
	return &eventUsecase{eventRepo: eventRepo, gormTransactioner: gormTransactioner, eventOptionRepo: eventOptionRepo, participantRepo: participantRepo, paymentRepo: paymentRepo, paymentRecordRepo: paymentRecordRepo, venueRepo: venueRepo}
}

func (u *eventUsecase) GetByID(ctx context.Context, id string) (*model.Event, error) {
	return u.eventRepo.FindByID(ctx, id)
}

func (u *eventUsecase) GetByShareToken(ctx context.Context, token string) (*model.Event, error) {
	return u.eventRepo.FindByShareToken(ctx, token)
}

func (u *eventUsecase) List(ctx context.Context) ([]*model.Event, error) {
	return u.eventRepo.List(ctx)
}

func (u *eventUsecase) ListForDashboard(ctx context.Context) ([]*model.EventSummary, error) {
	events, err := u.eventRepo.ListWithCreator(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]*model.EventSummary, len(events))
	for i, event := range events {
		summary := &model.EventSummary{
			Event: *event,
		}

		// Populate status-specific fields
		switch event.Status {
		case model.EventStatusVoting:
			// Get total votes for this event
			options, err := u.eventOptionRepo.FindByEventIDWithVoteCount(ctx, event.ID, nil)
			if err == nil {
				var totalVotes int64
				for _, opt := range options {
					totalVotes += opt.VoteCount
				}
				summary.TotalVotes = totalVotes
			}

		case model.EventStatusConfirmed, model.EventStatusOpen:
			// Get participant count
			count, err := u.participantRepo.CountByEventID(ctx, event.ID)
			if err == nil {
				summary.ParticipantCount = count
			}
			// Get chosen option details
			if event.ChosenOptionID != nil {
				option, err := u.eventOptionRepo.FindByID(ctx, *event.ChosenOptionID)
				if err == nil && option != nil {
					summary.EventDate = option.Date.Format("2006-01-02")
					summary.EventTime = option.StartTime + " - " + option.EndTime
					// Get venue name
					venue, err := u.venueRepo.FindByID(ctx, option.VenueID)
					if err == nil && venue != nil {
						summary.VenueName = venue.Name
					}
				}
			}

		case model.EventStatusPaymentOpen, model.EventStatusCompleted:
			// Get participant count
			count, err := u.participantRepo.CountByEventID(ctx, event.ID)
			if err == nil {
				summary.ParticipantCount = count
			}
			// Get chosen option details
			if event.ChosenOptionID != nil {
				option, err := u.eventOptionRepo.FindByID(ctx, *event.ChosenOptionID)
				if err == nil && option != nil {
					summary.EventDate = option.Date.Format("2006-01-02")
					summary.EventTime = option.StartTime + " - " + option.EndTime
					venue, err := u.venueRepo.FindByID(ctx, option.VenueID)
					if err == nil && venue != nil {
						summary.VenueName = venue.Name
					}
				}
			}
			// Get payment record counts
			payment, err := u.paymentRepo.FindByEventID(ctx, event.ID)
			if err == nil && payment != nil {
				records, err := u.paymentRecordRepo.FindByPaymentID(ctx, payment.ID)
				if err == nil {
					var pending, claimed, confirmed int64
					for _, record := range records {
						switch record.Status {
						case model.PaymentRecordStatusPending:
							pending++
						case model.PaymentRecordStatusClaimed:
							claimed++
						case model.PaymentRecordStatusConfirmed:
							confirmed++
						}
					}
					summary.PendingCount = pending
					summary.ClaimedCount = claimed
					summary.ConfirmedCount = confirmed
				}
			}
		}

		summaries[i] = summary
	}

	return summaries, nil
}

func (u *eventUsecase) Create(ctx context.Context, userID string, req *model.CreateEventRequest) (*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"req": utils.Dump(req),
	})
	event := &model.Event{
		ID:             uuid.New().String(),
		CreatedBy:      userID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         model.EventStatusVoting,
		ShareToken:     generateShareToken(),
		PlayerCap:      req.PlayerCap,
		VotingDeadline: req.VotingDeadline,
	}

	tx := u.gormTransactioner.Begin(ctx)
	if err := u.eventRepo.CreateWithTx(ctx, tx, event); err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(req.CreateEventOptionRequests) > 0 {
		options := make([]*model.EventOption, len(req.CreateEventOptionRequests))
		for i := range req.CreateEventOptionRequests {
			options[i] = &model.EventOption{
				ID:        uuid.New().String(),
				EventID:   event.ID,
				VenueID:   req.CreateEventOptionRequests[i].VenueID,
				Date:      req.CreateEventOptionRequests[i].Date,
				StartTime: req.CreateEventOptionRequests[i].StartTime,
				EndTime:   req.CreateEventOptionRequests[i].EndTime,
			}
		}
		if err := u.eventOptionRepo.BulkCreateWithTx(ctx, tx, options); err != nil {
			logger.Error(err)
			tx.Rollback()
			return nil, err
		}
	}

	// Add creator as a participant automatically
	participant := &model.Participant{
		ID:       uuid.New().String(),
		EventID:  event.ID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	if err := u.participantRepo.CreateWithTx(ctx, tx, participant); err != nil {
		logger.Error(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return event, nil
}

func (u *eventUsecase) UpdateStatus(ctx context.Context, id string, status model.EventStatus) error {
	return u.eventRepo.UpdateStatus(ctx, id, status)
}

func (u *eventUsecase) UpdateChosenOption(ctx context.Context, id string, optionID string) error {
	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"id":       id,
		"optionID": optionID,
	})

	if err := u.eventRepo.UpdateChosenOption(ctx, id, optionID); err != nil {
		logger.Error(err)
		return err
	}

	return u.eventRepo.UpdateStatus(ctx, id, model.EventStatusConfirmed)
}

func (u *eventUsecase) Delete(ctx context.Context, id string) error {
	return u.eventRepo.Delete(ctx, id)
}

func generateShareToken() string {
	b := make([]byte, 15)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:10]
}
