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

func (r *paymentRepo) UpdateBaseSplit(ctx context.Context, id string, baseSplit int) error {
	if err := r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Update("base_split", baseSplit).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"id":        id,
			"baseSplit": baseSplit,
		}).Error(err)
		return fmt.Errorf("failed to update base split: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdateBaseSplitWithTx(ctx context.Context, tx *gorm.DB, id string, baseSplit int) error {
	if err := tx.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Update("base_split", baseSplit).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"id":        id,
			"baseSplit": baseSplit,
		}).Error(err)
		return fmt.Errorf("failed to update base split: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdateTotals(ctx context.Context, id string, totalCost, baseSplit, taxAmount int) error {
	if err := r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"total_cost": totalCost,
		"base_split": baseSplit,
		"tax_amount": taxAmount,
	}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"id":        id,
			"totalCost": totalCost,
			"baseSplit": baseSplit,
			"taxAmount": taxAmount,
		}).Error(err)
		return fmt.Errorf("failed to update payment totals: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdateTotalsWithTx(ctx context.Context, tx *gorm.DB, id string, totalCost, baseSplit, taxAmount int) error {
	if err := tx.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"total_cost": totalCost,
		"base_split": baseSplit,
		"tax_amount": taxAmount,
	}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"id":        id,
			"totalCost": totalCost,
			"baseSplit": baseSplit,
			"taxAmount": taxAmount,
		}).Error(err)
		return fmt.Errorf("failed to update payment totals: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdateConfigWithTx(ctx context.Context, tx *gorm.DB, id string, paymentType model.PaymentType, totalCost, baseSplit, taxAmount int) error {
	if err := tx.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"type":       paymentType,
		"total_cost": totalCost,
		"base_split": baseSplit,
		"tax_amount": taxAmount,
	}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"id":        id,
			"type":      paymentType,
			"totalCost": totalCost,
			"baseSplit": baseSplit,
			"taxAmount": taxAmount,
		}).Error(err)
		return fmt.Errorf("failed to update payment config: %w", err)
	}
	return nil
}

func (r *paymentRepo) UpdatePaymentInfo(ctx context.Context, id string, paymentMethodID *string, paymentInfo string, paymentImageURL string) error {
	updates := map[string]interface{}{
		"payment_method_id": paymentMethodID,
		"payment_info":      paymentInfo,
		"payment_image_url": paymentImageURL,
	}
	if err := r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":             utils.DumpIncomingContext(ctx),
			"id":              id,
			"paymentMethodID": paymentMethodID,
			"paymentInfo":     paymentInfo,
			"paymentImageURL": paymentImageURL,
		}).Error(err)
		return fmt.Errorf("failed to update payment info: %w", err)
	}
	return nil
}
