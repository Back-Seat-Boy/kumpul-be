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
	if err := r.db.WithContext(ctx).
		Preload("Participant").
		Preload("Participant.User").
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claimed_at DESC")
		}).
		Where("payment_id = ?", paymentID).
		Find(&records).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list payment records: %w", err)
	}
	return records, nil
}

func (r *paymentRecordRepo) FindByPaymentIDAndParticipantID(ctx context.Context, paymentID, participantID string) (*model.PaymentRecord, error) {
	logger := log.WithFields(log.Fields{
		"context":       utils.DumpIncomingContext(ctx),
		"paymentID":     paymentID,
		"participantID": participantID,
	})

	var record model.PaymentRecord
	if err := r.db.WithContext(ctx).
		Preload("Participant").
		Preload("Participant.User").
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claimed_at DESC")
		}).
		Where("payment_id = ? AND participant_id = ?", paymentID, participantID).
		First(&record).Error; err != nil {
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

func (r *paymentRecordRepo) DeleteByPaymentIDAndParticipantID(ctx context.Context, paymentID, participantID string) error {
	if err := r.db.WithContext(ctx).Where("payment_id = ? AND participant_id = ?", paymentID, participantID).Delete(&model.PaymentRecord{}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":           utils.DumpIncomingContext(ctx),
			"paymentID":     paymentID,
			"participantID": participantID,
		}).Error(err)
		return fmt.Errorf("failed to delete payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) DeleteByPaymentIDAndParticipantIDWithTx(ctx context.Context, tx *gorm.DB, paymentID, participantID string) error {
	if err := tx.WithContext(ctx).Where("payment_id = ? AND participant_id = ?", paymentID, participantID).Delete(&model.PaymentRecord{}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":           utils.DumpIncomingContext(ctx),
			"paymentID":     paymentID,
			"participantID": participantID,
		}).Error(err)
		return fmt.Errorf("failed to delete payment record: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) UpdateSplitAmountByPaymentID(ctx context.Context, paymentID string, splitAmount int) error {
	if err := r.db.WithContext(ctx).Model(&model.PaymentRecord{}).
		Where("payment_id = ?", paymentID).
		Update("amount", splitAmount).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"paymentID":   paymentID,
			"splitAmount": splitAmount,
		}).Error(err)
		return fmt.Errorf("failed to update payment record split amount: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) UpdateSplitAmountByPaymentIDWithTx(ctx context.Context, tx *gorm.DB, paymentID string, splitAmount int) error {
	if err := tx.WithContext(ctx).Model(&model.PaymentRecord{}).
		Where("payment_id = ?", paymentID).
		Update("amount", splitAmount).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"paymentID":   paymentID,
			"splitAmount": splitAmount,
		}).Error(err)
		return fmt.Errorf("failed to update payment record split amount: %w", err)
	}
	return nil
}

func (r *paymentRecordRepo) CountConfirmedByPaymentID(ctx context.Context, paymentID string) (int64, error) {
	logger := log.WithFields(log.Fields{
		"context":   utils.DumpIncomingContext(ctx),
		"paymentID": paymentID,
	})

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.PaymentRecord{}).
		Where("payment_id = ? AND status = ?", paymentID, model.PaymentRecordStatusConfirmed).
		Count(&count).Error; err != nil {
		logger.Error(err)
		return 0, fmt.Errorf("failed to count confirmed payment records: %w", err)
	}
	return count, nil
}
