package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
)

type paymentUsecase struct {
	paymentRepo       model.PaymentRepository
	paymentRecordRepo model.PaymentRecordRepository
	participantRepo   model.ParticipantRepository
	eventRepo         model.EventRepository
}

func NewPaymentUsecase(paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, participantRepo model.ParticipantRepository, eventRepo model.EventRepository) model.PaymentUsecase {
	return &paymentUsecase{
		paymentRepo:       paymentRepo,
		paymentRecordRepo: paymentRecordRepo,
		participantRepo:   participantRepo,
		eventRepo:         eventRepo,
	}
}

func (u *paymentUsecase) GetByEventID(ctx context.Context, eventID string) (*model.Payment, error) {
	return u.paymentRepo.FindByEventID(ctx, eventID)
}

func (u *paymentUsecase) Create(ctx context.Context, eventID string, req *model.CreatePaymentRequest) (*model.Payment, error) {
	logger := log.WithFields(log.Fields{
		"context":     utils.DumpIncomingContext(ctx),
		"eventID":     eventID,
		"totalCost":   req.TotalCost,
		"paymentInfo": req.PaymentInfo,
	})

	// Check event status - can only create payment when status is "open"
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if event.Status != model.EventStatusOpen {
		return nil, model.ErrEventNotOpenForJoining
	}

	participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if participantCount == 0 {
		return nil, model.ErrNoParticipantsInEvent
	}

	splitAmount := req.TotalCost / int(participantCount)

	payment := &model.Payment{
		ID:          uuid.New().String(),
		EventID:     eventID,
		TotalCost:   req.TotalCost,
		SplitAmount: splitAmount,
		PaymentInfo: req.PaymentInfo,
		CreatedAt:   time.Now(),
	}

	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		logger.Error(err)
		return nil, err
	}

	// Create payment records for all current participants
	participants, err := u.participantRepo.FindByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for _, p := range participants {
		record := &model.PaymentRecord{
			ID:        uuid.New().String(),
			PaymentID: payment.ID,
			UserID:    p.UserID,
			Status:    model.PaymentRecordStatusPending,
		}

		// Creator's payment record is auto-confirmed (they hold the money)
		if p.UserID == event.CreatedBy {
			record.Status = model.PaymentRecordStatusConfirmed
			record.PaidAmount = splitAmount
			now := time.Now()
			record.ConfirmedAt = &now
		}

		if err := u.paymentRecordRepo.Create(ctx, record); err != nil {
			logger.Error(err)
			return nil, err
		}
	}

	return payment, nil
}

func (u *paymentUsecase) RecalculateSplitAmount(ctx context.Context, eventID string) error {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if err == model.ErrPaymentNotFound {
			// No payment yet, nothing to recalculate
			return nil
		}
		logger.Error(err)
		return err
	}

	participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return err
	}

	if participantCount == 0 {
		return model.ErrNoParticipantsInEvent
	}

	newSplitAmount := payment.TotalCost / int(participantCount)

	if err := u.paymentRepo.UpdateSplitAmount(ctx, payment.ID, newSplitAmount); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
