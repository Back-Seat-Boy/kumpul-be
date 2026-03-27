package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type paymentRecordUsecase struct {
	recordRepo      model.PaymentRecordRepository
	paymentRepo     model.PaymentRepository
	eventRepo       model.EventRepository
	participantRepo model.ParticipantRepository
}

func NewPaymentRecordUsecase(recordRepo model.PaymentRecordRepository, paymentRepo model.PaymentRepository, eventRepo model.EventRepository, participantRepo model.ParticipantRepository) model.PaymentRecordUsecase {
	return &paymentRecordUsecase{recordRepo: recordRepo, paymentRepo: paymentRepo, eventRepo: eventRepo, participantRepo: participantRepo}
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
		diff := record.PaidAmount - payment.SplitAmount
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
			actionAmount = payment.SplitAmount
			diff = -payment.SplitAmount
		}

		perPersonStatus = append(perPersonStatus, model.ParticipantPaymentStatus{
			UserID:       record.UserID,
			UserName:     record.User.Name,
			Status:       status,
			PaidAmount:   record.PaidAmount,
			CurrentSplit: payment.SplitAmount,
			Difference:   diff,
			Action:       action,
			ActionAmount: actionAmount,
		})
	}

	// Total should collect is based on current split amount x total participants
	totalShouldCollect := int(participantCount) * payment.SplitAmount
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
	if event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotInPaymentPhase
	}

	record, err := u.recordRepo.FindByPaymentIDAndUserID(ctx, paymentID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	record.Status = model.PaymentRecordStatusClaimed
	record.ProofImageURL = req.ProofImageURL
	now := time.Now()
	record.ClaimedAt = &now

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *paymentRecordUsecase) Confirm(ctx context.Context, paymentID string, userID string) error {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"paymentID": paymentID,
		"userID":    userID,
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
	if event.Status != model.EventStatusPaymentOpen {
		return model.ErrEventNotInPaymentPhase
	}

	record, err := u.recordRepo.FindByPaymentIDAndUserID(ctx, paymentID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	record.Status = model.PaymentRecordStatusConfirmed
	record.PaidAmount = payment.SplitAmount
	now := time.Now()
	record.ConfirmedAt = &now

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *paymentRecordUsecase) AdjustPayment(ctx context.Context, paymentID string, userID string, requesterID string, req *model.AdjustPaymentRequest) error {
	logger := log.WithFields(log.Fields{
		"context":          utils.DumpIncomingContext(ctx),
		"paymentID":        paymentID,
		"userID":           userID,
		"requesterID":      requesterID,
		"adjustmentAmount": req.AdjustmentAmount,
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

	if event.CreatedBy != requesterID {
		return model.ErrForbidden
	}

	record, err := u.recordRepo.FindByPaymentIDAndUserID(ctx, paymentID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Only confirmed payments can be adjusted
	if record.Status != model.PaymentRecordStatusConfirmed {
		return model.ErrPaymentRecordNotConfirmed
	}

	// Apply adjustment
	record.PaidAmount += req.AdjustmentAmount
	if record.PaidAmount < 0 {
		record.PaidAmount = 0
	}

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
