package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type paymentMethodRepo struct {
	db *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) model.PaymentMethodRepository {
	return &paymentMethodRepo{db: db}
}

func (r *paymentMethodRepo) FindByID(ctx context.Context, id string) (*model.PaymentMethod, error) {
	var paymentMethod model.PaymentMethod
	if err := r.db.WithContext(ctx).First(&paymentMethod, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentMethodNotFound
		}
		return nil, fmt.Errorf("failed to find payment method: %w", err)
	}
	return &paymentMethod, nil
}

func (r *paymentMethodRepo) FindByIDAndUserID(ctx context.Context, id, userID string) (*model.PaymentMethod, error) {
	var paymentMethod model.PaymentMethod
	if err := r.db.WithContext(ctx).First(&paymentMethod, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentMethodNotFound
		}
		return nil, fmt.Errorf("failed to find payment method: %w", err)
	}
	return &paymentMethod, nil
}

func (r *paymentMethodRepo) ListByUserID(ctx context.Context, userID string) ([]*model.PaymentMethod, error) {
	var paymentMethods []*model.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}
	return paymentMethods, nil
}

func (r *paymentMethodRepo) Create(ctx context.Context, paymentMethod *model.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Create(paymentMethod).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":           utils.DumpIncomingContext(ctx),
			"paymentMethod": utils.Dump(paymentMethod),
		}).Error(err)
		return fmt.Errorf("failed to create payment method: %w", err)
	}
	return nil
}

func (r *paymentMethodRepo) Update(ctx context.Context, paymentMethod *model.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Model(&model.PaymentMethod{}).
		Where("id = ?", paymentMethod.ID).
		Updates(map[string]interface{}{
			"label":        paymentMethod.Label,
			"payment_info": paymentMethod.PaymentInfo,
			"image_url":    paymentMethod.ImageURL,
		}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":           utils.DumpIncomingContext(ctx),
			"paymentMethod": utils.Dump(paymentMethod),
		}).Error(err)
		return fmt.Errorf("failed to update payment method: %w", err)
	}
	return nil
}

func (r *paymentMethodRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.PaymentMethod{}, "id = ?", id).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
			"id":  id,
		}).Error(err)
		return fmt.Errorf("failed to delete payment method: %w", err)
	}
	return nil
}
