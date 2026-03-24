package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type paymentRecordUsecase struct {
	recordRepo  model.PaymentRecordRepository
	paymentRepo model.PaymentRepository
	eventRepo   model.EventRepository
}

func NewPaymentRecordUsecase(recordRepo model.PaymentRecordRepository, paymentRepo model.PaymentRepository, eventRepo model.EventRepository) model.PaymentRecordUsecase {
	return &paymentRecordUsecase{recordRepo: recordRepo, paymentRepo: paymentRepo, eventRepo: eventRepo}
}

func (u *paymentRecordUsecase) GetByPaymentID(ctx context.Context, paymentID string) ([]*model.PaymentRecord, error) {
	return u.recordRepo.FindByPaymentID(ctx, paymentID)
}

func (u *paymentRecordUsecase) Claim(ctx context.Context, paymentID string, userID string, req *model.ClaimPaymentRequest) error {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"paymentID":     paymentID,
		"userID":        userID,
		"proofImageURL": req.ProofImageURL,
	})

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

	record, err := u.recordRepo.FindByPaymentIDAndUserID(ctx, paymentID, userID)
	if err != nil {
		logger.Error(err)
		return err
	}

	record.Status = model.PaymentRecordStatusConfirmed
	now := time.Now()
	record.ConfirmedAt = &now

	if err := u.recordRepo.Update(ctx, record); err != nil {
		logger.Error(err)
		return err
	}

	// check if all payments are confirmed and auto complete event
	if err := u.checkAndCompleteEvent(ctx, paymentID); err != nil {
		logger.WithError(err).Warn("Failed to check event completion status")
	}

	return nil
}

// checkAndCompleteEvent checks if all participants have confirmed payments and marks event as completed
func (u *paymentRecordUsecase) checkAndCompleteEvent(ctx context.Context, paymentID string) error {
	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}

	event, err := u.eventRepo.FindByID(ctx, payment.EventID)
	if err != nil {
		return err
	}

	if event.Status != model.EventStatusPaymentOpen {
		return nil
	}

	records, err := u.recordRepo.FindByPaymentID(ctx, paymentID)
	if err != nil {
		return err
	}

	allConfirmed := true
	for _, record := range records {
		if record.Status != model.PaymentRecordStatusConfirmed {
			allConfirmed = false
			break
		}
	}

	if allConfirmed && len(records) > 0 {
		if err := u.eventRepo.UpdateStatus(ctx, event.ID, model.EventStatusCompleted); err != nil {
			return err
		}
	}

	return nil
}
