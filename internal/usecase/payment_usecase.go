package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type paymentUsecase struct {
	paymentRepo       model.PaymentRepository
	paymentRecordRepo model.PaymentRecordRepository
	splitBillRepo     model.SplitBillRepository
	participantRepo   model.ParticipantRepository
	eventRepo         model.EventRepository
	paymentRecordUC   model.PaymentRecordUsecase
	gormTransactioner model.GormTransactioner
}

func NewPaymentUsecase(paymentRepo model.PaymentRepository, paymentRecordRepo model.PaymentRecordRepository, splitBillRepo model.SplitBillRepository, participantRepo model.ParticipantRepository, eventRepo model.EventRepository, paymentRecordUC model.PaymentRecordUsecase, gormTransactioner model.GormTransactioner) model.PaymentUsecase {
	return &paymentUsecase{
		paymentRepo:       paymentRepo,
		paymentRecordRepo: paymentRecordRepo,
		splitBillRepo:     splitBillRepo,
		participantRepo:   participantRepo,
		eventRepo:         eventRepo,
		paymentRecordUC:   paymentRecordUC,
		gormTransactioner: gormTransactioner,
	}
}

func (u *paymentUsecase) GetByEventID(ctx context.Context, eventID string) (*model.Payment, error) {
	return u.paymentRepo.FindByEventID(ctx, eventID)
}

func (u *paymentUsecase) GetSplitBillDetails(ctx context.Context, paymentID string) (*model.SplitBillDetails, error) {
	payment, err := u.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment.Type != model.PaymentTypeSplitBill {
		return nil, nil
	}

	items, err := u.splitBillRepo.FindByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	participantMap := make(map[string]*model.Participant)
	for _, item := range items {
		for idx := range item.Assignments {
			assignment := &item.Assignments[idx]
			participant := assignment.Participant
			participant.SetDerivedFields()
			participantCopy := participant
			participantMap[assignment.ParticipantID] = &participantCopy
		}
	}

	computation, err := computeSplitBill(items, participantMap, payment.TaxAmount)
	if err != nil {
		return nil, err
	}

	participantIDs := make([]string, 0, len(participantMap))
	for participantID := range participantMap {
		participantIDs = append(participantIDs, participantID)
	}
	sort.Strings(participantIDs)

	result := &model.SplitBillDetails{
		TaxAmount:     payment.TaxAmount,
		ItemsSubtotal: computation.ItemsSubtotal,
		GrandTotal:    computation.GrandTotal,
		Items:         computation.ItemDetails,
	}
	for _, participantID := range participantIDs {
		result.Participants = append(result.Participants, model.SplitBillParticipantDetail{
			ParticipantID: participantID,
			DisplayName:   participantDisplayName(participantMap[participantID]),
			Subtotal:      computation.SubtotalByParticipant[participantID],
			TaxAmount:     computation.TaxByParticipant[participantID],
			Total:         computation.TotalByParticipant[participantID],
		})
	}

	return result, nil
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

	splitBillItems, amountByParticipant, totalCost, baseSplit, taxAmount, err := u.preparePaymentConfig(ctx, eventID, paymentType, req.TotalCost, req.PerPersonAmount, req.TaxAmount, req.SplitBillItems)
	if err != nil {
		return nil, err
	}

	payment := &model.Payment{
		ID:          uuid.New().String(),
		EventID:     eventID,
		TotalCost:   totalCost,
		BaseSplit:   baseSplit,
		TaxAmount:   taxAmount,
		Type:        paymentType,
		PaymentInfo: req.PaymentInfo,
		CreatedAt:   time.Now(),
	}

	tx := u.gormTransactioner.Begin(ctx)
	if err := tx.WithContext(ctx).Create(payment).Error; err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Create payment records for all current participants
	participants, err := u.participantRepo.FindByEventID(ctx, eventID)
	if err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return nil, err
	}

	for _, p := range participants {
		amount := payment.BaseSplit
		if payment.Type == model.PaymentTypeSplitBill {
			amount = amountByParticipant[p.ID]
		}

		record := &model.PaymentRecord{
			ID:            uuid.New().String(),
			PaymentID:     payment.ID,
			ParticipantID: p.ID,
			Amount:        amount,
			Status:        model.PaymentRecordStatusPending,
		}

		// Creator's payment record is auto-confirmed (they hold the money)
		if p.UserID.Valid && p.UserID.String == event.CreatedBy {
			record.Status = model.PaymentRecordStatusConfirmed
			record.PaidAmount = amount
			now := time.Now()
			record.ConfirmedAt = &now
		}

		if err := u.paymentRecordRepo.CreateWithTx(ctx, tx, record); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return nil, err
		}
	}

	if payment.Type == model.PaymentTypeSplitBill {
		if err := u.persistSplitBillConfig(ctx, tx, payment.ID, splitBillItems); err != nil {
			logger.Error(err)
			u.gormTransactioner.Rollback(tx)
			return nil, err
		}
	}

	if err := u.gormTransactioner.Commit(tx); err != nil {
		logger.Error(err)
		return nil, err
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

	if err := u.paymentRepo.UpdateTotals(ctx, payment.ID, newTotalCost, newBaseSplit, payment.TaxAmount); err != nil {
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

	paymentType, err := normalizePaymentType(req.Type)
	if err != nil {
		return nil, err
	}

	splitBillItems, amountByParticipant, totalCost, baseSplit, taxAmount, err := u.preparePaymentConfig(ctx, eventID, paymentType, req.TotalCost, req.PerPersonAmount, req.TaxAmount, req.SplitBillItems)
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
	if err := u.paymentRepo.UpdateConfigWithTx(ctx, tx, payment.ID, paymentType, totalCost, baseSplit, taxAmount); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return nil, err
	}

	if err := u.splitBillRepo.DeleteByPaymentIDWithTx(ctx, tx, payment.ID); err != nil {
		logger.Error(err)
		u.gormTransactioner.Rollback(tx)
		return nil, err
	}

	for _, record := range records {
		record.Amount = baseSplit
		if paymentType == model.PaymentTypeSplitBill {
			record.Amount = amountByParticipant[record.ParticipantID]
		}
		if record.Participant.UserID.Valid && record.Participant.UserID.String == event.CreatedBy {
			record.Status = model.PaymentRecordStatusConfirmed
			record.PaidAmount = record.Amount
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

	if paymentType == model.PaymentTypeSplitBill {
		if err := u.persistSplitBillConfig(ctx, tx, payment.ID, splitBillItems); err != nil {
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
	payment.TaxAmount = taxAmount
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
	case model.PaymentTypeSplitBill:
		return model.PaymentTypeSplitBill, nil
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
	case model.PaymentTypeSplitBill:
		if totalCost < 0 {
			return 0, 0, fmt.Errorf("total_cost must be greater than or equal to 0 for split_bill payment type")
		}
		return totalCost, totalCost / participantCount, nil
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
	case model.PaymentTypeSplitBill:
		return payment.TotalCost, payment.TotalCost / participantCount, 0, nil
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

func (u *paymentUsecase) preparePaymentConfig(ctx context.Context, eventID string, paymentType model.PaymentType, totalCost, perPersonAmount, taxAmount int, inputs []model.SplitBillItemInput) ([]*model.SplitBillItem, map[string]int, int, int, int, error) {
	participants, err := u.participantRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, nil, 0, 0, 0, err
	}

	participantMap := make(map[string]*model.Participant, len(participants))
	for _, participant := range participants {
		participant.SetDerivedFields()
		participantMap[participant.ID] = participant
	}

	if paymentType != model.PaymentTypeSplitBill {
		finalTotalCost, baseSplit, err := buildPaymentAmounts(paymentType, totalCost, perPersonAmount, len(participants))
		return nil, nil, finalTotalCost, baseSplit, 0, err
	}

	if err := validateSplitBillInputs(inputs, participantMap, taxAmount); err != nil {
		return nil, nil, 0, 0, 0, err
	}

	items := buildSplitBillModels("", inputs)
	computation, err := computeSplitBill(items, participantMap, taxAmount)
	if err != nil {
		return nil, nil, 0, 0, 0, err
	}

	baseSplit := 0
	if len(participants) > 0 {
		baseSplit = computation.GrandTotal / len(participants)
	}

	return items, computation.TotalByParticipant, computation.GrandTotal, baseSplit, taxAmount, nil
}

func (u *paymentUsecase) persistSplitBillConfig(ctx context.Context, tx *gorm.DB, paymentID string, items []*model.SplitBillItem) error {
	if err := u.splitBillRepo.DeleteByPaymentIDWithTx(ctx, tx, paymentID); err != nil {
		return err
	}

	for _, item := range items {
		item.PaymentID = paymentID
		if err := u.splitBillRepo.CreateItemWithTx(ctx, tx, item); err != nil {
			return err
		}
		for idx := range item.Assignments {
			assignment := item.Assignments[idx]
			assignment.ItemID = item.ID
			if err := u.splitBillRepo.CreateAssignmentWithTx(ctx, tx, &assignment); err != nil {
				return err
			}
		}
	}

	return nil
}
