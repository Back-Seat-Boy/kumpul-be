package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type paymentRecordUsecase struct {
	recordRepo      model.PaymentRecordRepository
	claimRepo       model.PaymentClaimRepository
	paymentRepo     model.PaymentRepository
	eventRepo       model.EventRepository
	participantRepo model.ParticipantRepository
}

func NewPaymentRecordUsecase(recordRepo model.PaymentRecordRepository, claimRepo model.PaymentClaimRepository, paymentRepo model.PaymentRepository, eventRepo model.EventRepository, participantRepo model.ParticipantRepository) model.PaymentRecordUsecase {
	return &paymentRecordUsecase{recordRepo: recordRepo, claimRepo: claimRepo, paymentRepo: paymentRepo, eventRepo: eventRepo, participantRepo: participantRepo}
}

func (u *paymentRecordUsecase) GetByPaymentID(ctx context.Context, paymentID string) (*model.PaymentRecordsWithSummary, error) {
	records, err := u.recordRepo.FindByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Get payment info for split amount
	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Count participants
	participantCount, err := u.participantRepo.CountByEventID(ctx, payment.EventID)
	if err != nil {
		participantCount = int64(len(records))
	}

	// Calculate status counts and total collected (based on actual paid_amount)
	var numConfirmed, numClaimed, numPending int64
	var totalCollected int
	var perPersonStatus []model.ParticipantPaymentStatus

	for _, record := range records {
		status := string(record.Status)
		switch record.Status {
		case model.PaymentRecordStatusConfirmed:
			numConfirmed++
			totalCollected += record.PaidAmount
		case model.PaymentRecordStatusClaimed:
			numClaimed++
		case model.PaymentRecordStatusPending:
			numPending++
		}

		// Calculate per-person status
		diff := record.PaidAmount - record.Amount
		action := ""
		actionAmount := 0

		switch record.Status {
		case model.PaymentRecordStatusConfirmed:
			if diff > 0 {
				action = "receive_refund"
				actionAmount = diff
			} else if diff < 0 {
				action = "pay_more"
				actionAmount = -diff
			} else {
				action = "no_action"
				actionAmount = 0
			}
		case model.PaymentRecordStatusPending:
			action = "pay_full"
			actionAmount = record.Amount
			diff = -record.Amount
		}

		perPersonStatus = append(perPersonStatus, model.ParticipantPaymentStatus{
			ParticipantID: record.ParticipantID,
			DisplayName:   paymentRecordDisplayName(record),
			Status:        status,
			Amount:        record.Amount,
			PaidAmount:    record.PaidAmount,
			Difference:    diff,
			Action:        action,
			ActionAmount:  actionAmount,
		})
	}

	var totalShouldCollect int
	for _, record := range records {
		totalShouldCollect += record.Amount
	}
	balance := totalCollected - totalShouldCollect

	return &model.PaymentRecordsWithSummary{
		Records:            records,
		NumParticipants:    participantCount,
		NumConfirmed:       numConfirmed,
		NumClaimed:         numClaimed,
		NumPending:         numPending,
		TotalCollected:     totalCollected,
		TotalShouldCollect: totalShouldCollect,
		Balance:            balance,
		PerPersonStatus:    perPersonStatus,
	}, nil
}

func (u *paymentRecordUsecase) Claim(ctx context.Context, paymentID string, userID string, req *model.ClaimPaymentRequest) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"paymentID":     paymentID,
		"userID":        userID,
		"proofImageURL": req.ProofImageURL,
	})

	// Check event status - can only claim when status is "payment_open"
	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		logger.Error(err)
		return err
	}
	event, err := u.eventRepo.FindByID(ctx, payment.EventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}
	if event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotInPaymentPhase
	}

	participant, err := u.participantRepo.FindByEventIDAndUserID(ctx, payment.EventID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	record, err := u.recordRepo.FindByPaymentIDAndParticipantID(ctx, paymentID, participant.ID)
	if err != nil {
		logger.Error(err)
		return err
	}

	claimAmount := record.Amount - record.PaidAmount
	if claimAmount < 0 {
		claimAmount = 0
	}

	now := time.Now()
	claim := &model.PaymentClaim{
		ID:              uuid.New().String(),
		PaymentRecordID: record.ID,
		ParticipantID:   participant.ID,
		ClaimedAmount:   claimAmount,
		ProofImageURL:   req.ProofImageURL,
		Status:          model.PaymentClaimStatusClaimed,
		ClaimedAt:       now,
	}
	if err := u.claimRepo.Create(ctx, claim); err != nil {
		logger.Error(err)
		return err
	}

	record.Status = model.PaymentRecordStatusClaimed
	record.ProofImageURL = req.ProofImageURL
	record.ClaimedAt = &now

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *paymentRecordUsecase) Confirm(ctx context.Context, paymentID string, participantID string, req *model.ConfirmPaymentRequest) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"paymentID":     paymentID,
		"participantID": participantID,
	})

	// Get current payment to record the split amount at confirmation time
	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Check event status - can only confirm when status is "payment_open"
	event, err := u.eventRepo.FindByID(ctx, payment.EventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}
	if event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotInPaymentPhase
	}

	record, err := u.recordRepo.FindByPaymentIDAndParticipantID(ctx, paymentID, participantID)
	if err != nil {
		logger.Error(err)
		return err
	}

	targetAmount := record.Amount
	if req != nil && req.Amount != nil {
		record.Amount = *req.Amount
		targetAmount = *req.Amount
	}

	claim, err := u.claimRepo.FindLatestClaimedByPaymentRecordID(ctx, record.ID)
	if err != nil {
		logger.Error(err)
		return err
	}

	confirmedAmount := 0
	if claim != nil {
		confirmedAmount = claim.ClaimedAmount
	} else {
		confirmedAmount = targetAmount - record.PaidAmount
		if confirmedAmount < 0 {
			confirmedAmount = 0
		}
	}

	now := time.Now()
	if claim != nil {
		claim.Status = model.PaymentClaimStatusConfirmed
		claim.ConfirmedAt = &now
		if req != nil && req.ProofImageURL != "" {
			claim.ProofImageURL = req.ProofImageURL
		}
		if req != nil && req.Note != "" {
			claim.Note = appendPaymentNote(claim.Note, req.Note)
		}
		if err := u.claimRepo.Update(ctx, claim); err != nil {
			logger.Error(err)
			return err
		}
	} else {
		directClaim := &model.PaymentClaim{
			ID:              uuid.New().String(),
			PaymentRecordID: record.ID,
			ParticipantID:   participantID,
			ClaimedAmount:   confirmedAmount,
			Status:          model.PaymentClaimStatusConfirmed,
			ClaimedAt:       now,
			ConfirmedAt:     &now,
		}
		if req != nil {
			directClaim.ProofImageURL = req.ProofImageURL
			directClaim.Note = req.Note
		}
		if err := u.claimRepo.Create(ctx, directClaim); err != nil {
			logger.Error(err)
			return err
		}
	}

	record.Status = model.PaymentRecordStatusConfirmed
	record.PaidAmount += confirmedAmount
	if req != nil {
		record.Note = appendPaymentNote(record.Note, req.Note)
		if req.ProofImageURL != "" {
			record.ProofImageURL = req.ProofImageURL
		}
	}
	record.ConfirmedAt = &now

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *paymentRecordUsecase) AdjustPayment(ctx context.Context, paymentID string, participantID string, requesterID string, req *model.AdjustPaymentRequest) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"paymentID":     paymentID,
		"participantID": participantID,
		"requesterID":   requesterID,
		"amount":        req.Amount,
	})

	// Get payment to check event
	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Get event to verify requester is the creator and check status
	event, err := u.eventRepo.FindByID(ctx, payment.EventID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Can only adjust payments when status is "payment_open"
	if event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotInPaymentPhase
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}

	if event.CreatedBy != requesterID {
		return model.ErrForbidden
	}
	if payment.Type == model.PaymentTypeSplitBill {
		return model.ErrSplitBillManualAdjustBlocked
	}

	record, err := u.recordRepo.FindByPaymentIDAndParticipantID(ctx, paymentID, participantID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Apply adjustment regardless of current record status.
	record.Amount = req.Amount
	record.Note = appendPaymentNote(record.Note, req.Note)
	if req.ProofImageURL != "" {
		record.ProofImageURL = req.ProofImageURL
	}

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *paymentRecordUsecase) ChargeAll(ctx context.Context, paymentID string, requesterID string, req *model.ChargeAllRequest) error {
	logger := log.WithFields(log.Fields{
		"context":     utils.DumpIncomingContext(ctx),
		"paymentID":   paymentID,
		"requesterID": requesterID,
		"amount":      req.Amount,
	})

	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		logger.Error(err)
		return err
	}

	event, err := u.eventRepo.FindByID(ctx, payment.EventID)
	if err != nil {
		logger.Error(err)
		return err
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return err
	}
	if event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotInPaymentPhase
	}
	if event.CreatedBy != requesterID {
		return model.ErrForbidden
	}
	if payment.Type == model.PaymentTypeSplitBill {
		return model.ErrSplitBillChargeAllBlocked
	}

	records, err := u.recordRepo.FindByPaymentID(ctx, paymentID)
	if err != nil {
		logger.Error(err)
		return err
	}

	for _, record := range records {
		record.Amount += req.Amount
		record.Note = appendPaymentNote(record.Note, req.Note)
		if err := u.recordRepo.Update(ctx, record); err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func paymentRecordDisplayName(record *model.PaymentRecord) string {
	if record == nil {
		return ""
	}
	if record.Participant.GuestName != "" {
		return record.Participant.GuestName
	}
	return record.Participant.User.Name
}

func appendPaymentNote(existing, incoming string) string {
	if incoming == "" {
		return existing
	}
	if existing == "" {
		return incoming
	}
	return existing + "\n" + incoming
}
