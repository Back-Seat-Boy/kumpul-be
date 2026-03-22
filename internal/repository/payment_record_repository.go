package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type paymentRecordRepo struct {
	db *gorm.DB
}

func NewPaymentRecordRepository(db *gorm.DB) model.PaymentRecordRepository {
	return &paymentRecordRepo{db: db}
}

func (r *paymentRecordRepo) FindByPaymentID(ctx context.Context, paymentID string) ([]*model.PaymentRecord, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"paymentID": paymentID,
	})

	var records []*model.PaymentRecord
	if err := r.db.WithContext(ctx).Preload("User").Where("payment_id = ?", paymentID).Find(&records).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list payment records: %w", err)
	}
	return records, nil
}

func (r *paymentRecordRepo) FindByPaymentIDAndUserID(ctx context.Context, paymentID, userID string) (*model.PaymentRecord, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"paymentID": paymentID,
		"userID":    userID,
	})

	var record model.PaymentRecord
	if err := r.db.WithContext(ctx).Where("payment_id = ? AND user_id = ?", paymentID, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrPaymentRecordNotFound
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find payment record: %w", err)
	}
	return &record, nil
}

func (r *paymentRecordRepo) Create(ctx context.Context, record *model.PaymentRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"record": utils.Dump(record),
		}).Error(err)
		return fmt.Errorf("failed to create payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, record *model.PaymentRecord) error {
	if err := tx.WithContext(ctx).Create(record).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"record": utils.Dump(record),
		}).Error(err)
		return fmt.Errorf("failed to create payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) Update(ctx context.Context, record *model.PaymentRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"record": utils.Dump(record),
		}).Error(err)
		return fmt.Errorf("failed to update payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) DeleteByPaymentIDAndUserID(ctx context.Context, paymentID, userID string) error {
	if err := r.db.WithContext(ctx).Where("payment_id = ? AND user_id = ?", paymentID, userID).Delete(&model.PaymentRecord{}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"paymentID": paymentID,
			"userID":    userID,
		}).Error(err)
		return fmt.Errorf("failed to delete payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) DeleteByPaymentIDAndUserIDWithTx(ctx context.Context, tx *gorm.DB, paymentID, userID string) error {
	if err := tx.WithContext(ctx).Where("payment_id = ? AND user_id = ?", paymentID, userID).Delete(&model.PaymentRecord{}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":       utils.DumpIncomingContext(ctx),
			"paymentID": paymentID,
			"userID":    userID,
		}).Error(err)
		return fmt.Errorf("failed to delete payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) UpdateSplitAmountByPaymentID(ctx context.Context, paymentID string, splitAmount int) error {
	// This updates all records' split amount (stored in Payment table, not PaymentRecord)
	// This method is a no-op since split_amount is in Payment table
	// Kept for interface consistency
	return nil
}
