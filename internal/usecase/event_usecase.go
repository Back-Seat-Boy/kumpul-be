package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/guregu/null"
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

func (u *eventUsecase) ListCreatedByUser(ctx context.Context, userID string) ([]*model.EventSummary, error) {
	events, err := u.eventRepo.FindByCreatedBy(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.buildEventSummaries(ctx, events)
}

func (u *eventUsecase) ListParticipatedByUser(ctx context.Context, userID string) ([]*model.EventSummary, error) {
	events, err := u.eventRepo.FindByParticipantUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.buildEventSummaries(ctx, events)
}

func (u *eventUsecase) ListForDashboard(ctx context.Context, req *model.ListEventsRequest) (*model.ListEventsResponse, error) {
	return u.listEventsWithSummaries(ctx, req)
}

func (u *eventUsecase) ListPublic(ctx context.Context, req *model.ListEventsRequest) (*model.ListEventsResponse, error) {
	if req == nil {
		req = &model.ListEventsRequest{}
	}
	req.Filter.Visibility = model.EventVisibilityPublic
	return u.listEventsWithSummaries(ctx, req)
}

func (u *eventUsecase) listEventsWithSummaries(ctx context.Context, req *model.ListEventsRequest) (*model.ListEventsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	events, total, err := u.eventRepo.ListPaginated(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &model.ListEventsResponse{
			Events:  []*model.EventSummary{},
			Total:   total,
			HasMore: false,
		}, nil
	}

	summaries, err := u.buildEventSummaries(ctx, events)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	hasMore := false

	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		nextCursor = lastEvent.ID

		if req.Mode == model.PaginationModeCursor {
			hasMore = int64(len(events)) == int64(req.Limit) && total > int64(len(events))
		} else {
			offset := req.Page * req.Limit
			hasMore = int64(offset) < total
		}
	}

	return &model.ListEventsResponse{
		Events:     summaries,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (u *eventUsecase) buildEventSummaries(ctx context.Context, events []*model.Event) ([]*model.EventSummary, error) {
	if len(events) == 0 {
		return []*model.EventSummary{}, nil
	}

	eventIDs := make([]string, len(events))
	chosenOptionIDs := make([]string, 0)
	for i, e := range events {
		eventIDs[i] = e.ID
		if e.ChosenOptionID != nil {
			chosenOptionIDs = append(chosenOptionIDs, *e.ChosenOptionID)
		}
	}

	var (
		voteCounts        map[string]int64
		participantCounts map[string]int64
		optionDetails     map[string]*model.ChosenOptionDetails
		paymentCounts     map[string]*model.PaymentRecordCounts
	)

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		voteCounts, err = u.eventRepo.BulkFetchVoteCounts(ctx, eventIDs)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		participantCounts, err = u.eventRepo.BulkFetchParticipantCounts(ctx, eventIDs)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		optionDetails, err = u.eventRepo.BulkFetchChosenOptionDetails(ctx, chosenOptionIDs)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		paymentCounts, err = u.eventRepo.BulkFetchPaymentRecordCounts(ctx, eventIDs)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		log.WithError(err).Warn("Failed to fetch bulk data for events")
	}

	summaries := make([]*model.EventSummary, len(events))
	for i, event := range events {
		summary := &model.EventSummary{
			Event: *event,
		}

		switch event.Status {
		case model.EventStatusVoting:
			if count, ok := voteCounts[event.ID]; ok {
				summary.TotalVotes = count
			}

		case model.EventStatusConfirmed, model.EventStatusOpen:
			if count, ok := participantCounts[event.ID]; ok {
				summary.ParticipantCount = count
			}
			if event.ChosenOptionID != nil {
				if details, ok := optionDetails[*event.ChosenOptionID]; ok {
					summary.EventDate = details.Date
					summary.EventTime = details.StartTime + " - " + details.EndTime
					summary.VenueName = details.VenueName
				}
			}

		case model.EventStatusPaymentOpen, model.EventStatusCompleted:
			if count, ok := participantCounts[event.ID]; ok {
				summary.ParticipantCount = count
			}
			if event.ChosenOptionID != nil {
				if details, ok := optionDetails[*event.ChosenOptionID]; ok {
					summary.EventDate = details.Date
					summary.EventTime = details.StartTime + " - " + details.EndTime
					summary.VenueName = details.VenueName
				}
			}
			if counts, ok := paymentCounts[event.ID]; ok {
				summary.PendingCount = counts.Pending
				summary.ClaimedCount = counts.Claimed
				summary.ConfirmedCount = counts.Confirmed
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
	visibility := req.Visibility
	if visibility == "" {
		visibility = model.EventVisibilityInviteOnly
	}

	event := &model.Event{
		ID:             uuid.New().String(),
		CreatedBy:      userID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         model.EventStatusVoting,
		Visibility:     visibility,
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

	participant := &model.Participant{
		ID:       uuid.New().String(),
		EventID:  event.ID,
		UserID:   null.StringFrom(userID),
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
	event, err := u.eventRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !isValidEventStatusTransition(event.Status, status) {
		return model.ErrForbidden
	}
	return u.eventRepo.UpdateStatus(ctx, id, status)
}

func (u *eventUsecase) UpdateChosenOption(ctx context.Context, id string, optionID string) error {
	logger := log.WithFields(log.Fields{
		"context":  utils.DumpIncomingContext(ctx),
		"id":       id,
		"optionID": optionID,
	})

	event, err := u.eventRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}

	option, err := u.eventOptionRepo.FindByID(ctx, optionID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if option.EventID != id {
		return model.ErrEventOptionNotFound
	}

	voters, err := u.eventOptionRepo.FindVotersByOptionID(ctx, optionID)
	if err != nil {
		logger.Error(err)
		return err
	}

	existingParticipants, err := u.participantRepo.FindByEventID(ctx, id)
	if err != nil {
		logger.Error(err)
		return err
	}

	existingUserIDs := make(map[string]struct{}, len(existingParticipants))
	for _, participant := range existingParticipants {
		if participant.UserID.Valid && participant.UserID.String != "" {
			existingUserIDs[participant.UserID.String] = struct{}{}
		}
	}

	missingParticipants := make([]*model.Participant, 0, len(voters))
	for _, voter := range voters {
		if _, exists := existingUserIDs[voter.UserID]; exists {
			continue
		}
		missingParticipants = append(missingParticipants, &model.Participant{
			ID:       uuid.New().String(),
			EventID:  id,
			UserID:   null.StringFrom(voter.UserID),
			JoinedAt: time.Now(),
		})
		existingUserIDs[voter.UserID] = struct{}{}
	}

	tx := u.gormTransactioner.Begin(ctx)
	if err := u.eventRepo.UpdateChosenOptionWithTx(ctx, tx, id, optionID); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return err
	}

	for _, participant := range missingParticipants {
		if err := u.participantRepo.CreateWithTx(ctx, tx, participant); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
	}

	if err := u.eventRepo.UpdateStatusWithTx(ctx, tx, id, model.EventStatusConfirmed); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return err
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *eventUsecase) Delete(ctx context.Context, id string) error {
	return u.eventRepo.Delete(ctx, id)
}

// CheckAndCompleteEvent checks if all participants have confirmed payments and marks event as completed
func (u *eventUsecase) CheckAndCompleteEvent(ctx context.Context, eventID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	// Get event
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Only check if event is in payment_open status
	if event.Status != model.EventStatusPaymentOpen {
		return nil
	}

	// Get payment
	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			return nil
		}
		logger.Error(err)
		return err
	}

	// Get all payment records
	records, err := u.paymentRecordRepo.FindByPaymentID(ctx, payment.ID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Check if all records are confirmed
	allConfirmed := true
	for _, record := range records {
		if record.Status != model.PaymentRecordStatusConfirmed {
			allConfirmed = false
			break
		}
	}

	// If all confirmed, mark event as completed
	if allConfirmed && len(records) > 0 {
		if err := u.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusCompleted); err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func generateShareToken() string {
	b := make([]byte, 15)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:10]
}
