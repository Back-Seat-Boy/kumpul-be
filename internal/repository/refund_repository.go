package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type refundRepo struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) model.RefundRepository {
	return &refundRepo{db: db}
}

func (r *refundRepo) FindByID(ctx context.Context, id string) (*model.Refund, error) {
	var refund model.Refund
	if err := r.db.WithContext(ctx).First(&refund, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.ErrRefundNotFound
		}
		return nil, fmt.Errorf("failed to find refund: %w", err)
	}
	return &refund, nil
}

func (r *refundRepo) FindByEventID(ctx context.Context, eventID string) ([]*model.Refund, error) {
	var refunds []*model.Refund
	if err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}
	return refunds, nil
}

func (r *refundRepo) FindByUserID(ctx context.Context, userID string) ([]*model.Refund, error) {
	var refunds []*model.Refund
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}
	return refunds, nil
}

func (r *refundRepo) Create(ctx context.Context, refund *model.Refund) error {
	if err := r.db.WithContext(ctx).Create(refund).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"refund": utils.Dump(refund),
		}).Error(err)
		return fmt.Errorf("failed to create refund: %w", err)
	}
	return nil
}

func (r *refundRepo) Update(ctx context.Context, refund *model.Refund) error {
	if err := r.db.WithContext(ctx).Model(&model.Refund{}).
		Where("id = ?", refund.ID).
		Updates(map[string]interface{}{
			"status":                      refund.Status,
			"recipient_payment_method_id": refund.RecipientPaymentMethodID,
			"recipient_payment_info":      refund.RecipientPaymentInfo,
			"recipient_payment_image_url": refund.RecipientPaymentImageURL,
			"recipient_note":              refund.RecipientNote,
			"sent_proof_image_url":        refund.SentProofImageURL,
			"sent_note":                   refund.SentNote,
			"sent_at":                     refund.SentAt,
			"received_at":                 refund.ReceivedAt,
			"updated_at":                  refund.UpdatedAt,
		}).Error; err != nil {
		log.WithFields(log.Fields{
			"ctx":    utils.DumpIncomingContext(ctx),
			"refund": utils.Dump(refund),
		}).Error(err)
		return fmt.Errorf("failed to update refund: %w", err)
	}
	return nil
}
