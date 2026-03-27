package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type paymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) model.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) FindByID(ctx context.Context, id string) (*model.Payment, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"id":      id,
	})

	var payment model.Payment
	if err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	return &payment, nil
}

func (r *paymentRepo) FindByEventID(ctx context.Context, eventID string) (*model.Payment, error) {
	logger := log.WithFields(log.Fields{
		"context": utils.DumpIncomingContext(ctx),
		"eventID": eventID,
	})

	var payment model.Payment
	if err := r.db.WithContext(ctx).First(&payment, "event_id = ?", eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	return &payment, nil
}

func (r *paymentRepo) Create(ctx context.Context, payment *model.Payment) error {
	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":     utils.DumpIncomingContext(ctx),
			"payment": utils.Dump(payment),
		}).Error(err)
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdateSplitAmount(ctx context.Context, id string, splitAmount int) error {
	if err := r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Update("split_amount", splitAmount).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"id":          id,
			"splitAmount": splitAmount,
		}).Error(err)
		return fmt.Errorf("failed to update split amount: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdateSplitAmountWithTx(ctx context.Context, tx *gorm.DB, id string, splitAmount int) error {
	if err := tx.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Update("split_amount", splitAmount).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"id":          id,
			"splitAmount": splitAmount,
		}).Error(err)
		return fmt.Errorf("failed to update split amount: %w", err)
	}
	return nil
}
