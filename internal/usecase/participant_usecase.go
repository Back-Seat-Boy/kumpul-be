package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type participantUsecase struct {
	participantRepo   model.ParticipantRepository
	paymentRepo       model.PaymentRepository
	paymentRecordRepo model.PaymentRecordRepository
	eventRepo         model.EventRepository
	gormTransactioner model.GormTransactioner
}

func NewParticipantUsecase(participantRepo model.ParticipantRepository, paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, eventRepo model.EventRepository, gormTransactioner model.GormTransactioner) model.ParticipantUsecase {
	return &participantUsecase{
		participantRepo:   participantRepo,
		paymentRepo:       paymentRepo,
		paymentRecordRepo: paymentRecordRepo,
		eventRepo:         eventRepo,
		gormTransactioner: gormTransactioner,
	}
}

func (u *participantUsecase) ListByEvent(ctx context.Context, eventID string) ([]*model.Participant, error) {
	return u.participantRepo.FindByEventID(ctx, eventID)
}

func (u *participantUsecase) Join(ctx context.Context, eventID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	// Check event status - can join when status is "open" or "payment_open"
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if event.Status == model.EventStatusCompleted {
		return model.ErrEventAlreadyCompleted
	}
	if event.Status != model.EventStatusOpen && event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotOpenForJoining
	}

	_, err = u.participantRepo.FindByEventIDAndUserID(ctx, eventID, userID)
	if err == nil {
		return model.ErrAlreadyJoined
	}

	participant := &model.Participant{
		ID:       uuid.New().String(),
		EventID:  eventID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := u.participantRepo.Create(ctx, participant); err != nil {
		logger.Error(err)
		return err
	}

	// Handle payment record creation if payment exists
	if err := u.HandlePaymentOnJoin(ctx, eventID, userID); err != nil {
		logger.Error(err)
		// Don't fail the join if payment handling fails, just log it
	}

	return nil
}

func (u *participantUsecase) Leave(ctx context.Context, eventID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	// Check event status - cannot leave if event is completed
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if event.Status == model.EventStatusCompleted {
		return model.ErrEventAlreadyCompleted
	}

	participant, err := u.participantRepo.FindByEventIDAndUserID(ctx, eventID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Handle payment record deletion before removing participant
	if err := u.HandlePaymentOnLeave(ctx, eventID, userID); err != nil {
		logger.Error(err)
		// Don't fail the leave if payment handling fails, just log it
	}

	if err := u.participantRepo.Delete(ctx, participant.ID); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) RemoveParticipant(ctx context.Context, eventID string, participantUserID string, requesterUserID string) (*model.RemoveParticipantResult, error) {
	logger := log.WithFields(log.Fields{
		"context":           utils.DumpIncomingContext(ctx),
		"eventID":           eventID,
		"participantUserID": participantUserID,
		"requesterUserID":   requesterUserID,
	})

	// Get event to verify requester is the creator and check status
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// Cannot remove participants if event is completed
	if event.Status == model.EventStatusCompleted {
		return nil, model.ErrEventAlreadyCompleted
	}

	if event.CreatedBy != requesterUserID {
		return nil, model.ErrForbidden
	}

	// Cannot remove creator
	if participantUserID == event.CreatedBy {
		return nil, model.ErrForbidden
	}

	participant, err := u.participantRepo.FindByEventIDAndUserID(ctx, eventID, participantUserID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// Get current payment info before removal for calculation
	var oldSplitAmount int
	var totalCost int
	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil && err != model.ErrPaymentNotFound {
		logger.Error(err)
		return nil, err
	}
	if err == nil {
		oldSplitAmount = payment.SplitAmount
		totalCost = payment.TotalCost
	}

	// Count total participants before removal
	totalParticipants, _ := u.participantRepo.CountByEventID(ctx, eventID)

	// Handle payment record deletion before removing participant
	if err := u.HandlePaymentOnLeave(ctx, eventID, participantUserID); err != nil {
		logger.Error(err)
		// Don't fail if payment handling fails, just log it
	}

	if err := u.participantRepo.Delete(ctx, participant.ID); err != nil {
		logger.Error(err)
		return nil, err
	}

	// Calculate the financial impact
	result := &model.RemoveParticipantResult{
		RemovedParticipantID: participantUserID,
		OldSplitAmount:       oldSplitAmount,
		NumRemaining:         totalParticipants - 1,
	}

	if payment != nil && totalParticipants > 1 {
		// Calculate new split amount
		newSplitAmount := totalCost / int(totalParticipants-1)
		result.NewSplitAmount = newSplitAmount

		// Get all payment records to calculate actual total collected and per-person impact
		allRecords, _ := u.paymentRecordRepo.FindByPaymentID(ctx, payment.ID)
		var totalCollected int
		var numPaidAfter int64
		var impacts []model.RemoveParticipantImpact

		for _, record := range allRecords {
			// Skip the removed participant
			if record.UserID == participantUserID {
				continue
			}
			if record.Status == model.PaymentRecordStatusConfirmed {
				numPaidAfter++
				totalCollected += record.PaidAmount

				// Calculate impact for each confirmed participant
				diff := record.PaidAmount - newSplitAmount
				action := ""
				actionAmount := 0

				if diff > 0 {
					action = "receive_refund"
					actionAmount = diff
				} else if diff < 0 {
					action = "pay_more"
					actionAmount = -diff
				} else {
					action = "no_action"
				}

				impacts = append(impacts, model.RemoveParticipantImpact{
					UserID:       record.UserID,
					UserName:     record.User.Name,
					PaidAmount:   record.PaidAmount,
					NewSplit:     newSplitAmount,
					Difference:   diff,
					Action:       action,
					ActionAmount: actionAmount,
				})
			}
		}

		result.NumPaidParticipants = numPaidAfter
		result.TotalCollected = totalCollected
		result.TotalShouldCollect = int(totalParticipants-1) * newSplitAmount
		result.Difference = result.TotalCollected - result.TotalShouldCollect
		result.Impacts = impacts
	}

	return result, nil
}

func (u *participantUsecase) GetParticipantCount(ctx context.Context, eventID string) (int64, error) {
	return u.participantRepo.CountByEventID(ctx, eventID)
}

func (u *participantUsecase) HandlePaymentOnJoin(ctx context.Context, eventID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	// Check if payment exists for this event
	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			// No payment yet, nothing to do
			return nil
		}
		logger.Error(err)
		return err
	}

	// Get event to check if joining user is creator
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Create payment record for new participant
	record := &model.PaymentRecord{
		ID:        uuid.New().String(),
		PaymentID: payment.ID,
		UserID:    userID,
		Status:    model.PaymentRecordStatusPending,
	}

	// Creator's payment record is auto-confirmed (they hold the money)
	if userID == event.CreatedBy {
		record.Status = model.PaymentRecordStatusConfirmed
		record.PaidAmount = payment.SplitAmount
		now := time.Now()
		record.ConfirmedAt = &now
	}

	tx := u.gormTransactioner.Begin(ctx)
	if err := u.paymentRecordRepo.CreateWithTx(ctx, tx, record); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return err
	}

	// Recalculate split amount for all participants
	participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return err
	}

	if participantCount > 0 {
		newSplitAmount := payment.TotalCost / int(participantCount)
		if err := u.paymentRepo.UpdateSplitAmountWithTx(ctx, tx, payment.ID, newSplitAmount); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) HandlePaymentOnLeave(ctx context.Context, eventID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	// Check if payment exists for this event
	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			// No payment yet, nothing to do
			return nil
		}
		logger.Error(err)
		return err
	}

	tx := u.gormTransactioner.Begin(ctx)
	// Delete payment record for leaving participant
	if err := u.paymentRecordRepo.DeleteByPaymentIDAndUserIDWithTx(ctx, tx, payment.ID, userID); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return err
	}

	// Recalculate split amount for remaining participants
	participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return err
	}

	// participantCount still includes the leaving participant at this point
	// so we subtract 1 for the calculation
	remainingCount := participantCount - 1
	if remainingCount > 0 {
		newSplitAmount := payment.TotalCost / int(remainingCount)
		if err := u.paymentRepo.UpdateSplitAmountWithTx(ctx, tx, payment.ID, newSplitAmount); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
