package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type participantUsecase struct {
	participantRepo   model.ParticipantRepository
	paymentRepo       model.PaymentRepository
	paymentRecordRepo model.PaymentRecordRepository
	splitBillRepo     model.SplitBillRepository
	eventRepo         model.EventRepository
	gormTransactioner model.GormTransactioner
}

func NewParticipantUsecase(participantRepo model.ParticipantRepository, paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, splitBillRepo model.SplitBillRepository, eventRepo model.EventRepository, gormTransactioner model.GormTransactioner) model.ParticipantUsecase {
	return &participantUsecase{
		participantRepo:   participantRepo,
		paymentRepo:       paymentRepo,
		paymentRecordRepo: paymentRecordRepo,
		splitBillRepo:     splitBillRepo,
		eventRepo:         eventRepo,
		gormTransactioner: gormTransactioner,
	}
}

func (u *participantUsecase) ListByEvent(ctx context.Context, eventID string) ([]*model.Participant, error) {
	return u.participantRepo.FindByEventID(ctx, eventID)
}

func (u *participantUsecase) Join(ctx context.Context, eventID string, userID string, viaShareLink bool) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
		"userID":  userID,
	})

	_, err := u.checkEventStatusForJoining(ctx, eventID, viaShareLink)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = u.participantRepo.FindByEventIDAndUserID(ctx, eventID, userID)
	if err == nil {
		return model.ErrAlreadyJoined
	}

	participant := &model.Participant{
		ID:       uuid.New().String(),
		EventID:  eventID,
		UserID:   null.StringFrom(userID),
		JoinedAt: time.Now(),
	}

	if err := u.participantRepo.Create(ctx, participant); err != nil {
		logger.Error(err)
		return err
	}

	// Handle payment record creation if payment exists
	if err := u.HandlePaymentOnJoin(ctx, eventID, participant.ID); err != nil {
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
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}

	participant, err := u.participantRepo.FindByEventIDAndUserID(ctx, eventID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Handle payment record deletion before removing participant
	if err := u.HandlePaymentOnLeave(ctx, eventID, participant.ID); err != nil {
		logger.Error(err)
		return err
	}

	if err := u.participantRepo.Delete(ctx, participant.ID); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) PreviewRemoveParticipant(ctx context.Context, eventID string, participantID string, requesterUserID string) (*model.RemoveParticipantResult, error) {
	event, participant, payment, totalParticipants, err := u.validateParticipantRemoval(ctx, eventID, participantID, requesterUserID)
	if err != nil {
		return nil, err
	}

	return u.buildRemoveParticipantResult(ctx, event, participant, payment, totalParticipants)
}

func (u *participantUsecase) RemoveParticipant(ctx context.Context, eventID string, participantID string, requesterUserID string) (*model.RemoveParticipantResult, error) {
	logger := log.WithFields(log.Fields{
		"context":         utils.DumpIncomingContext(ctx),
		"eventID":         eventID,
		"participantID":   participantID,
		"requesterUserID": requesterUserID,
	})

	event, participant, payment, totalParticipants, err := u.validateParticipantRemoval(ctx, eventID, participantID, requesterUserID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	result, err := u.buildRemoveParticipantResult(ctx, event, participant, payment, totalParticipants)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// Handle payment record deletion before removing participant
	if err := u.HandlePaymentOnLeave(ctx, eventID, participant.ID); err != nil {
		logger.Error(err)
		return nil, err
	}

	if err := u.participantRepo.Delete(ctx, participant.ID); err != nil {
		logger.Error(err)
		return nil, err
	}

	return result, nil
}

func (u *participantUsecase) GetParticipantCount(ctx context.Context, eventID string) (int64, error) {
	return u.participantRepo.CountByEventID(ctx, eventID)
}

func (u *participantUsecase) HandlePaymentOnJoin(ctx context.Context, eventID string, participantID string) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"eventID":       eventID,
		"participantID": participantID,
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

	participant, err := u.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Create payment record for new participant
	record := &model.PaymentRecord{
		ID:            uuid.New().String(),
		PaymentID:     payment.ID,
		ParticipantID: participantID,
		Amount:        payment.BaseSplit,
		Status:        model.PaymentRecordStatusPending,
	}
	if payment.Type == model.PaymentTypeSplitBill {
		record.Amount = 0
	}

	// Creator's payment record is auto-confirmed (they hold the money)
	if participant.UserID.Valid && participant.UserID.String == event.CreatedBy {
		record.Status = model.PaymentRecordStatusConfirmed
		record.PaidAmount = record.Amount
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
		newTotalCost, newBaseSplit, amountDelta, err := recalculatePaymentAmounts(payment, int(participantCount))
		if err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
		if err := u.paymentRepo.UpdateTotalsWithTx(ctx, tx, payment.ID, newTotalCost, newBaseSplit, payment.TaxAmount); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
		if payment.Type == model.PaymentTypeTotal {
			if err := u.paymentRecordRepo.ShiftAmountsByPaymentIDWithTx(ctx, tx, payment.ID, amountDelta); err != nil {
				logger.Error(err)
				u.gormTransactioner.Rollback(tx)
				return err
			}
		}
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) HandlePaymentOnLeave(ctx context.Context, eventID string, participantID string) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"eventID":       eventID,
		"participantID": participantID,
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
	if payment.Type == model.PaymentTypeSplitBill {
		hasAssignments, err := u.splitBillRepo.HasAssignmentsForParticipant(ctx, payment.ID, participantID)
		if err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
		if hasAssignments {
			u.gormTransactioner.Rollback(tx)
			return model.ErrSplitBillParticipantAssigned
		}
	}
	// Delete payment record for leaving participant
	if err := u.paymentRecordRepo.DeleteByPaymentIDAndParticipantIDWithTx(ctx, tx, payment.ID, participantID); err != nil {
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
		newTotalCost, newBaseSplit, amountDelta, err := recalculatePaymentAmounts(payment, int(remainingCount))
		if err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
		if err := u.paymentRepo.UpdateTotalsWithTx(ctx, tx, payment.ID, newTotalCost, newBaseSplit, payment.TaxAmount); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return err
		}
		if payment.Type == model.PaymentTypeTotal {
			if err := u.paymentRecordRepo.ShiftAmountsByPaymentIDWithTx(ctx, tx, payment.ID, amountDelta); err != nil {
				logger.Error(err)
				u.gormTransactioner.Rollback(tx)
				return err
			}
		}
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *participantUsecase) validateParticipantRemoval(ctx context.Context, eventID string, participantID string, requesterUserID string) (*model.Event, *model.Participant, *model.Payment, int64, error) {
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	if event.Status == model.EventStatusCompleted {
		return nil, nil, nil, 0, model.ErrEventAlreadyCompleted
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return nil, nil, nil, 0, err
	}
	if event.CreatedBy != requesterUserID {
		return nil, nil, nil, 0, model.ErrForbidden
	}

	participantByID, err := u.participantRepo.FindByID(ctx, participantID)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	if participantByID.UserID.Valid && participantByID.UserID.String == event.CreatedBy {
		return nil, nil, nil, 0, model.ErrForbidden
	}

	participant, err := u.participantRepo.FindByEventIDAndID(ctx, eventID, participantByID.ID)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil && err != model.ErrPaymentNotFound {
		return nil, nil, nil, 0, err
	}
	if err == model.ErrPaymentNotFound {
		payment = nil
	}
	if payment != nil && payment.Type == model.PaymentTypeSplitBill {
		hasAssignments, err := u.splitBillRepo.HasAssignmentsForParticipant(ctx, payment.ID, participantID)
		if err != nil {
			return nil, nil, nil, 0, err
		}
		if hasAssignments {
			return nil, nil, nil, 0, model.ErrSplitBillParticipantAssigned
		}
	}

	totalParticipants, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	return event, participant, payment, totalParticipants, nil
}

func (u *participantUsecase) buildRemoveParticipantResult(ctx context.Context, _ *model.Event, participant *model.Participant, payment *model.Payment, totalParticipants int64) (*model.RemoveParticipantResult, error) {
	result := &model.RemoveParticipantResult{
		RemovedParticipantID: participant.ID,
		NumRemaining:         totalParticipants - 1,
	}

	if payment == nil {
		return result, nil
	}

	result.OldSplitAmount = payment.BaseSplit

	record, err := u.paymentRecordRepo.FindByPaymentIDAndParticipantID(ctx, payment.ID, participant.ID)
	if err != nil && err != model.ErrPaymentRecordNotFound {
		return nil, err
	}
	if err == nil {
		result.RemovedPayment = buildRemovedParticipantPayment(record)
	}

	if totalParticipants <= 1 {
		return result, nil
	}

	newTotalCost, newSplitAmount, _, err := recalculatePaymentAmounts(payment, int(totalParticipants-1))
	if err != nil {
		return nil, err
	}
	result.NewSplitAmount = newSplitAmount

	allRecords, err := u.paymentRecordRepo.FindByPaymentID(ctx, payment.ID)
	if err != nil {
		return nil, err
	}

	var totalCollected int
	var numPaidAfter int64
	var impacts []model.RemoveParticipantImpact

	for _, record := range allRecords {
		if record.ParticipantID == participant.ID {
			continue
		}
		if record.Status != model.PaymentRecordStatusConfirmed {
			continue
		}

		numPaidAfter++
		totalCollected += record.PaidAmount

		targetAmount := newSplitAmount
		if payment.Type == model.PaymentTypeSplitBill {
			targetAmount = record.Amount
		}

		diff := record.PaidAmount - targetAmount
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
			ParticipantID: record.ParticipantID,
			DisplayName:   paymentRecordDisplayName(record),
			PaidAmount:    record.PaidAmount,
			NewSplit:      targetAmount,
			Difference:    diff,
			Action:        action,
			ActionAmount:  actionAmount,
		})
	}

	result.NumPaidParticipants = numPaidAfter
	result.TotalCollected = totalCollected
	result.TotalShouldCollect = newTotalCost
	if payment.Type == model.PaymentTypeSplitBill {
		result.TotalShouldCollect = payment.TotalCost
	}
	result.Difference = result.TotalCollected - result.TotalShouldCollect
	result.Impacts = impacts

	return result, nil
}

func buildRemovedParticipantPayment(record *model.PaymentRecord) *model.RemovedParticipantPayment {
	if record == nil {
		return nil
	}

	var pendingClaimedAmount int
	for _, claim := range record.Claims {
		if claim.Status == model.PaymentClaimStatusClaimed {
			pendingClaimedAmount += claim.ClaimedAmount
		}
	}

	return &model.RemovedParticipantPayment{
		ParticipantID:        record.ParticipantID,
		DisplayName:          paymentRecordDisplayName(record),
		Status:               string(record.Status),
		Amount:               record.Amount,
		PaidAmount:           record.PaidAmount,
		RefundAmount:         record.PaidAmount,
		PendingClaimedAmount: pendingClaimedAmount,
		Claims:               record.Claims,
	}
}

func (u *participantUsecase) JoinAsGuest(ctx context.Context, userID, eventID string, req *model.JoinAsGuestRequest, viaShareLink bool) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"userID":  userID,
		"eventID": eventID,
		"req":     utils.Dump(req),
	})

	// Check event status - can join when status is "open" or "payment_open"
	_, err := u.checkEventStatusForJoining(ctx, eventID, viaShareLink)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = u.participantRepo.FindByEventIDAndGuestName(ctx, eventID, req.GuestName)
	if err == nil {
		return model.ErrAlreadyJoined
	}

	participant := &model.Participant{
		ID:        uuid.New().String(),
		EventID:   eventID,
		GuestName: req.GuestName,
		AddedBy:   null.StringFrom(userID),
		JoinedAt:  time.Now(),
	}

	if err := u.participantRepo.Create(ctx, participant); err != nil {
		logger.Error(err)
		return err
	}

	// Handle payment record creation if payment exists
	if err := u.HandlePaymentOnJoin(ctx, eventID, participant.ID); err != nil {
		logger.Error(err)
		// Don't fail the join if payment handling fails, just log it
	}

	return nil
}

func (u *participantUsecase) checkEventStatusForJoining(ctx context.Context, eventID string, viaShareLink bool) (*model.Event, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if event.Status == model.EventStatusCompleted {
		return nil, model.ErrEventAlreadyCompleted
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return nil, err
	}
	if event.Status != model.EventStatusConfirmed && event.Status != model.EventStatusOpen && event.Status != model.EventStatusPaymentOpen {
		return nil, model.ErrEventNotOpenForJoining
	}
	if event.Visibility == model.EventVisibilityInviteOnly && !viaShareLink {
		return nil, model.ErrInviteOnlyRequiresLink
	}
	if event.PlayerCap != nil {
		participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		if participantCount >= int64(*event.PlayerCap) {
			return nil, model.ErrParticipantCapReached
		}
	}
	return event, nil
}
