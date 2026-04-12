package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
)

type refundUsecase struct {
	refundRepo        model.RefundRepository
	eventRepo         model.EventRepository
	paymentMethodRepo model.PaymentMethodRepository
}

func NewRefundUsecase(refundRepo model.RefundRepository, eventRepo model.EventRepository, paymentMethodRepo model.PaymentMethodRepository) model.RefundUsecase {
	return &refundUsecase{
		refundRepo:        refundRepo,
		eventRepo:         eventRepo,
		paymentMethodRepo: paymentMethodRepo,
	}
}

func (u *refundUsecase) CreateForRemovedParticipant(ctx context.Context, event *model.Event, payment *model.Payment, participant *model.Participant, record *model.PaymentRecord) (*model.Refund, error) {
	if event == nil || payment == nil || participant == nil || record == nil {
		return nil, nil
	}
	if record.PaidAmount <= 0 {
		return nil, nil
	}

	status := model.RefundStatusPendingInfo
	if !participant.UserID.Valid || participant.UserID.String == "" {
		status = model.RefundStatusReadyToSend
	}

	refund := &model.Refund{
		ID:                   uuid.New().String(),
		EventID:              event.ID,
		PaymentID:            payment.ID,
		RemovedParticipantID: participant.ID,
		DisplayName:          paymentRecordDisplayName(record),
		Amount:               record.PaidAmount,
		Status:               status,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if participant.UserID.Valid && participant.UserID.String != "" {
		userID := participant.UserID.String
		refund.UserID = &userID
	}

	if err := u.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}

func (u *refundUsecase) ListByEvent(ctx context.Context, eventID string, requesterID string) ([]*model.Refund, error) {
	event, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event.CreatedBy != requesterID {
		return nil, model.ErrForbidden
	}
	return u.refundRepo.FindByEventID(ctx, eventID)
}

func (u *refundUsecase) ListByUserID(ctx context.Context, userID string) ([]*model.Refund, error) {
	return u.refundRepo.FindByUserID(ctx, userID)
}

func (u *refundUsecase) UpdateDestination(ctx context.Context, id string, userID string, req *model.UpdateRefundDestinationRequest) (*model.Refund, error) {
	refund, err := u.refundRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if refund.UserID == nil || *refund.UserID != userID {
		return nil, model.ErrForbidden
	}
	if refund.Status != model.RefundStatusPendingInfo && refund.Status != model.RefundStatusReadyToSend {
		return nil, model.NewAppError(nil, http.StatusBadRequest, "refund destination can no longer be changed")
	}

	var resolvedMethodID *string
	paymentInfo := req.PaymentInfo
	paymentImageURL := req.PaymentImageURL
	if req.PaymentMethodID != "" {
		method, err := u.paymentMethodRepo.FindByIDAndUserID(ctx, req.PaymentMethodID, userID)
		if err != nil {
			return nil, err
		}
		resolvedMethodID = &method.ID
		if paymentInfo == "" {
			paymentInfo = method.PaymentInfo
		}
		if paymentImageURL == "" {
			paymentImageURL = method.ImageURL
		}
	}
	if paymentInfo == "" {
		return nil, model.NewAppError(nil, http.StatusBadRequest, "payment_info or payment_method_id is required")
	}

	refund.RecipientPaymentMethodID = resolvedMethodID
	refund.RecipientPaymentInfo = paymentInfo
	refund.RecipientPaymentImageURL = paymentImageURL
	refund.RecipientNote = req.Note
	refund.Status = model.RefundStatusReadyToSend
	refund.UpdatedAt = time.Now()

	if err := u.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}

func (u *refundUsecase) MarkSent(ctx context.Context, id string, requesterID string, req *model.SendRefundRequest) (*model.Refund, error) {
	refund, err := u.refundRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	event, err := u.eventRepo.FindByID(ctx, refund.EventID)
	if err != nil {
		return nil, err
	}
	if event.CreatedBy != requesterID {
		return nil, model.ErrForbidden
	}
	if refund.Status == model.RefundStatusPendingInfo {
		return nil, model.ErrRefundDestinationRequired
	}
	if refund.Status != model.RefundStatusReadyToSend && refund.Status != model.RefundStatusSent {
		return nil, model.NewAppError(nil, http.StatusBadRequest, "refund cannot be marked as sent")
	}

	now := time.Now()
	refund.Status = model.RefundStatusSent
	refund.SentProofImageURL = req.ProofImageURL
	refund.SentNote = req.Note
	refund.SentAt = &now
	refund.UpdatedAt = now
	if err := u.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}

func (u *refundUsecase) ConfirmReceipt(ctx context.Context, id string, userID string) (*model.Refund, error) {
	refund, err := u.refundRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if refund.UserID == nil || *refund.UserID != userID {
		return nil, model.ErrForbidden
	}
	if refund.Status != model.RefundStatusSent {
		return nil, model.NewAppError(nil, http.StatusBadRequest, "refund has not been marked as sent yet")
	}

	now := time.Now()
	refund.Status = model.RefundStatusReceived
	refund.ReceivedAt = &now
	refund.UpdatedAt = now
	if err := u.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}
