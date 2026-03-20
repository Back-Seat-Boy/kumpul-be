package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type paymentRecordUsecase struct {
	recordRepo model.PaymentRecordRepository
}

func NewPaymentRecordUsecase(recordRepo model.PaymentRecordRepository) model.PaymentRecordUsecase {
	return &paymentRecordUsecase{recordRepo: recordRepo}
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

	return nil
}
