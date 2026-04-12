package usecase

import (
	"context"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/google/uuid"
)

type paymentMethodUsecase struct {
	paymentMethodRepo model.PaymentMethodRepository
}

func NewPaymentMethodUsecase(paymentMethodRepo model.PaymentMethodRepository) model.PaymentMethodUsecase {
	return &paymentMethodUsecase{paymentMethodRepo: paymentMethodRepo}
}

func (u *paymentMethodUsecase) ListByUserID(ctx context.Context, userID string) ([]*model.PaymentMethod, error) {
	return u.paymentMethodRepo.ListByUserID(ctx, userID)
}

func (u *paymentMethodUsecase) Create(ctx context.Context, userID string, req *model.CreatePaymentMethodRequest) (*model.PaymentMethod, error) {
	paymentMethod := &model.PaymentMethod{
		ID:          uuid.New().String(),
		UserID:      userID,
		Label:       req.Label,
		PaymentInfo: req.PaymentInfo,
		ImageURL:    req.ImageURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := u.paymentMethodRepo.Create(ctx, paymentMethod); err != nil {
		return nil, err
	}
	return paymentMethod, nil
}

func (u *paymentMethodUsecase) Update(ctx context.Context, id string, userID string, req *model.UpdatePaymentMethodRequest) (*model.PaymentMethod, error) {
	paymentMethod, err := u.paymentMethodRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	paymentMethod.Label = req.Label
	paymentMethod.PaymentInfo = req.PaymentInfo
	paymentMethod.ImageURL = req.ImageURL
	paymentMethod.UpdatedAt = time.Now()
	if err := u.paymentMethodRepo.Update(ctx, paymentMethod); err != nil {
		return nil, err
	}
	return paymentMethod, nil
}

func (u *paymentMethodUsecase) Delete(ctx context.Context, id string, userID string) error {
	if _, err := u.paymentMethodRepo.FindByIDAndUserID(ctx, id, userID); err != nil {
		return err
	}
	return u.paymentMethodRepo.Delete(ctx, id)
}
