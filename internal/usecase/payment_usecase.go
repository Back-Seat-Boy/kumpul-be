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
	paymentRecordUC   model.PaymentRecordUsecase
}

func NewPaymentUsecase(paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, participantRepo model.ParticipantRepository, eventRepo model.EventRepository, paymentRecordUC model.PaymentRecordUsecase) model.PaymentUsecase {
	return &paymentUsecase{
		paymentRepo:       paymentRepo,
		paymentRecordRepo: paymentRecordRepo,
		participantRepo:   participantRepo,
		eventRepo:         eventRepo,
		paymentRecordUC:   paymentRecordUC,
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
	if err := ensureEventNotCancelled(event); err != nil {
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

	baseSplit := req.TotalCost / int(participantCount)

	payment := &model.Payment{
		ID:          uuid.New().String(),
		EventID:     eventID,
		TotalCost:   req.TotalCost,
		BaseSplit:   baseSplit,
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
			ID:            uuid.New().String(),
			PaymentID:     payment.ID,
			ParticipantID: p.ID,
			Amount:        payment.BaseSplit,
			Status:        model.PaymentRecordStatusPending,
		}

		// Creator's payment record is auto-confirmed (they hold the money)
		if p.UserID.Valid && p.UserID.String == event.CreatedBy {
			record.Status = model.PaymentRecordStatusConfirmed
			record.PaidAmount = payment.BaseSplit
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

	newBaseSplit := payment.TotalCost / int(participantCount)

	if err := u.paymentRepo.UpdateBaseSplit(ctx, payment.ID, newBaseSplit); err != nil {
		logger.Error(err)
		return err
	}
	if err := u.paymentRecordRepo.UpdateSplitAmountByPaymentID(ctx, payment.ID, newBaseSplit); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *paymentUsecase) UpdatePaymentInfo(ctx context.Context, eventID string, requesterID string, req *model.UpdatePaymentRequest) (*model.Payment, error) {
	logger := log.WithFields(log.Fields{
		"context":     utils.DumpIncomingContext(ctx),
		"eventID":     eventID,
		"requesterID": requesterID,
	})

	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if event.CreatedBy != requesterID {
		return nil, model.ErrForbidden
	}
	if err := ensureEventNotCancelled(event); err != nil {
		return nil, err
	}
	if event.Status != model.EventStatusPaymentOpen {
		return nil, model.ErrEventNotInPaymentPhase
	}

	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if err := u.paymentRepo.UpdatePaymentInfo(ctx, payment.ID, req.PaymentInfo); err != nil {
		logger.Error(err)
		return nil, err
	}
	payment.PaymentInfo = req.PaymentInfo
	return payment, nil
}

func (u *paymentUsecase) ChargeAll(ctx context.Context, eventID string, requesterID string, req *model.ChargeAllRequest) error {
	logger := log.WithFields(log.Fields{
		"context":     utils.DumpIncomingContext(ctx),
		"eventID":     eventID,
		"requesterID": requesterID,
		"amount":      req.Amount,
	})

	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return err
	}

	return u.paymentRecordUC.ChargeAll(ctx, payment.ID, requesterID, req)
}
