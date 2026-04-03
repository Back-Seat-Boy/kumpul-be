package usecase

import (
	"context"
	"fmt"
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
	gormTransactioner model.GormTransactioner
}

func NewPaymentUsecase(paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, participantRepo model.ParticipantRepository, eventRepo model.EventRepository, paymentRecordUC model.PaymentRecordUsecase, gormTransactioner model.GormTransactioner) model.PaymentUsecase {
	return &paymentUsecase{
		paymentRepo:       paymentRepo,
		paymentRecordRepo: paymentRecordRepo,
		participantRepo:   participantRepo,
		eventRepo:         eventRepo,
		paymentRecordUC:   paymentRecordUC,
		gormTransactioner: gormTransactioner,
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

	paymentType, err := normalizePaymentType(req.Type)
	if err != nil {
		return nil, err
	}

	totalCost, baseSplit, err := buildPaymentAmounts(paymentType, req.TotalCost, req.PerPersonAmount, int(participantCount))
	if err != nil {
		return nil, err
	}

	payment := &model.Payment{
		ID:          uuid.New().String(),
		EventID:     eventID,
		TotalCost:   totalCost,
		BaseSplit:   baseSplit,
		Type:        paymentType,
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

	newTotalCost, newBaseSplit, amountDelta, err := recalculatePaymentAmounts(payment, int(participantCount))
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := u.paymentRepo.UpdateTotals(ctx, payment.ID, newTotalCost, newBaseSplit); err != nil {
		logger.Error(err)
		return err
	}
	if payment.Type == model.PaymentTypeTotal {
		if err := u.paymentRecordRepo.ShiftAmountsByPaymentID(ctx, payment.ID, amountDelta); err != nil {
			logger.Error(err)
			return err
		}
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

func (u *paymentUsecase) UpdatePaymentConfig(ctx context.Context, eventID string, requesterID string, req *model.UpdatePaymentConfigRequest) (*model.Payment, error) {
	logger := log.WithFields(log.Fields{
		"context":     utils.DumpIncomingContext(ctx),
		"eventID":     eventID,
		"requesterID": requesterID,
		"req":         utils.Dump(req),
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
	if event.Status != model.EventStatusOpen && event.Status != model.EventStatusPaymentOpen {
		return nil, model.ErrEventNotInPaymentPhase
	}

	payment, err := u.paymentRepo.FindByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	participantCount, err := u.participantRepo.CountByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	paymentType, err := normalizePaymentType(req.Type)
	if err != nil {
		return nil, err
	}

	totalCost, baseSplit, err := buildPaymentAmounts(paymentType, req.TotalCost, req.PerPersonAmount, int(participantCount))
	if err != nil {
		return nil, err
	}

	records, err := u.paymentRecordRepo.FindByPaymentID(ctx, payment.ID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if !paymentConfigEditable(records, event.CreatedBy) {
		return nil, model.ErrPaymentConfigLocked
	}

	tx := u.gormTransactioner.Begin(ctx)
	if err := u.paymentRepo.UpdateConfigWithTx(ctx, tx, payment.ID, paymentType, totalCost, baseSplit); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return nil, err
	}

	for _, record := range records {
		record.Amount = baseSplit
		if record.Participant.UserID.Valid && record.Participant.UserID.String == event.CreatedBy {
			record.Status = model.PaymentRecordStatusConfirmed
			record.PaidAmount = baseSplit
			now := time.Now()
			record.ConfirmedAt = &now
		} else {
			record.Status = model.PaymentRecordStatusPending
			record.PaidAmount = 0
		}

		if err := u.paymentRecordRepo.UpdateWithTx(ctx, tx, record); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return nil, err
		}
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return nil, err
	}

	payment.Type = paymentType
	payment.TotalCost = totalCost
	payment.BaseSplit = baseSplit
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

func normalizePaymentType(raw string) (model.PaymentType, error) {
	switch model.PaymentType(raw) {
	case "":
		return model.PaymentTypeTotal, nil
	case model.PaymentTypeTotal:
		return model.PaymentTypeTotal, nil
	case model.PaymentTypePerPerson:
		return model.PaymentTypePerPerson, nil
	default:
		return "", fmt.Errorf("invalid payment type: %s", raw)
	}
}

func buildPaymentAmounts(paymentType model.PaymentType, totalCost, perPersonAmount, participantCount int) (int, int, error) {
	if participantCount <= 0 {
		return 0, 0, model.ErrNoParticipantsInEvent
	}

	switch paymentType {
	case model.PaymentTypeTotal:
		if totalCost <= 0 {
			return 0, 0, fmt.Errorf("total_cost must be greater than 0 for total payment type")
		}
		return totalCost, totalCost / participantCount, nil
	case model.PaymentTypePerPerson:
		if perPersonAmount <= 0 {
			return 0, 0, fmt.Errorf("per_person_amount must be greater than 0 for per_person payment type")
		}
		return perPersonAmount * participantCount, perPersonAmount, nil
	default:
		return 0, 0, fmt.Errorf("invalid payment type: %s", paymentType)
	}
}

func recalculatePaymentAmounts(payment *model.Payment, participantCount int) (int, int, int, error) {
	if payment == nil {
		return 0, 0, 0, model.ErrPaymentNotFound
	}
	if participantCount <= 0 {
		return 0, 0, 0, model.ErrNoParticipantsInEvent
	}

	switch payment.Type {
	case model.PaymentTypePerPerson:
		return payment.BaseSplit * participantCount, payment.BaseSplit, 0, nil
	case model.PaymentTypeTotal, "":
		newBaseSplit := payment.TotalCost / participantCount
		return payment.TotalCost, newBaseSplit, newBaseSplit - payment.BaseSplit, nil
	default:
		return 0, 0, 0, fmt.Errorf("invalid payment type: %s", payment.Type)
	}
}

func paymentConfigEditable(records []*model.PaymentRecord, creatorID string) bool {
	for _, record := range records {
		isCreator := record.Participant.UserID.Valid && record.Participant.UserID.String == creatorID
		if isCreator {
			continue
		}
		if record.PaidAmount > 0 || record.Status != model.PaymentRecordStatusPending || len(record.Claims) > 0 {
			return false
		}
	}
	return true
}
