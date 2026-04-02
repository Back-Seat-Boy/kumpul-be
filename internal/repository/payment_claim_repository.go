package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type paymentClaimRepo struct {
	db *gorm.DB
}

func NewPaymentClaimRepository(db *gorm.DB) model.PaymentClaimRepository {
	return &paymentClaimRepo{db: db}
}

func (r *paymentClaimRepo) Create(ctx context.Context, claim *model.PaymentClaim) error {
	if err := r.db.WithContext(ctx).Create(claim).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"claim": utils.Dump(claim),
		}).Error(err)
		return fmt.Errorf("failed to create payment claim: %w", err)
	}
	return nil
}

func (r *paymentClaimRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, claim *model.PaymentClaim) error {
	if err := tx.WithContext(ctx).Create(claim).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"claim": utils.Dump(claim),
		}).Error(err)
		return fmt.Errorf("failed to create payment claim: %w", err)
	}
	return nil
}

func (r *paymentClaimRepo) Update(ctx context.Context, claim *model.PaymentClaim) error {
	if err := r.db.WithContext(ctx).Save(claim).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":   utils.DumpIncomingContext(ctx),
			"claim": utils.Dump(claim),
		}).Error(err)
		return fmt.Errorf("failed to update payment claim: %w", err)
	}
	return nil
}

func (r *paymentClaimRepo) FindLatestClaimedByPaymentRecordID(ctx context.Context, paymentRecordID string) (*model.PaymentClaim, error) {
	logger := log.WithFields(log.Fields{
		"context":         utils.DumpIncomingContext(ctx),
		"paymentRecordID": paymentRecordID,
	})

	var claim model.PaymentClaim
	if err := r.db.WithContext(ctx).
		Where("payment_record_id = ? AND status = ?", paymentRecordID, model.PaymentClaimStatusClaimed).
		Order("claimed_at DESC").
		First(&claim).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error(err)
		return nil, fmt.Errorf("failed to find latest claimed payment claim: %w", err)
	}
	return &claim, nil
}

func (r *paymentClaimRepo) FindByPaymentRecordID(ctx context.Context, paymentRecordID string) ([]*model.PaymentClaim, error) {
	logger := log.WithFields(log.Fields{
		"context":         utils.DumpIncomingContext(ctx),
		"paymentRecordID": paymentRecordID,
	})

	var claims []*model.PaymentClaim
	if err := r.db.WithContext(ctx).
		Where("payment_record_id = ?", paymentRecordID).
		Order("claimed_at DESC").
		Find(&claims).Error; err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to list payment claims: %w", err)
	}
	return claims, nil
}
