package repository

import (
	"context"
	"fmt"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type splitBillRepo struct {
	db *gorm.DB
}

func NewSplitBillRepository(db *gorm.DB) model.SplitBillRepository {
	return &splitBillRepo{db: db}
}

func (r *splitBillRepo) FindByPaymentID(ctx context.Context, paymentID string) ([]*model.SplitBillItem, error) {
	var items []*model.SplitBillItem
	if err := r.db.WithContext(ctx).
		Preload("Assignments", func(db *gorm.DB) *gorm.DB {
			return db.Order("participant_id ASC")
		}).
		Preload("Assignments.Participant").
		Preload("Assignments.Participant.User").
		Where("payment_id = ?", paymentID).
		Order("created_at ASC, id ASC").
		Find(&items).Error; err != nil {
		log.WithFields(log.Fields{
			"context":   utils.DumpIncomingContext(ctx),
			"paymentID": paymentID,
		}).Error(err)
		return nil, fmt.Errorf("failed to find split bill items: %w", err)
	}
	return items, nil
}

func (r *splitBillRepo) DeleteByPaymentIDWithTx(ctx context.Context, tx *gorm.DB, paymentID string) error {
	if err := tx.WithContext(ctx).Where("payment_id = ?", paymentID).Delete(&model.SplitBillItem{}).Error; err != nil {
		log.WithFields(log.Fields{
			"context":   utils.DumpIncomingContext(ctx),
			"paymentID": paymentID,
		}).Error(err)
		return fmt.Errorf("failed to delete split bill items: %w", err)
	}
	return nil
}

func (r *splitBillRepo) CreateItemWithTx(ctx context.Context, tx *gorm.DB, item *model.SplitBillItem) error {
	if err := tx.WithContext(ctx).Omit("Assignments").Create(item).Error; err != nil {
		log.WithFields(log.Fields{
			"context": utils.DumpIncomingContext(ctx),
			"item":    utils.Dump(item),
		}).Error(err)
		return fmt.Errorf("failed to create split bill item: %w", err)
	}
	return nil
}

func (r *splitBillRepo) CreateAssignmentWithTx(ctx context.Context, tx *gorm.DB, assignment *model.SplitBillItemAssignment) error {
	if err := tx.WithContext(ctx).Create(assignment).Error; err != nil {
		log.WithFields(log.Fields{
			"context":    utils.DumpIncomingContext(ctx),
			"assignment": utils.Dump(assignment),
		}).Error(err)
		return fmt.Errorf("failed to create split bill assignment: %w", err)
	}
	return nil
}

func (r *splitBillRepo) HasAssignmentsForParticipant(ctx context.Context, paymentID, participantID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.SplitBillItemAssignment{}).
		Joins("JOIN split_bill_items ON split_bill_items.id = split_bill_item_assignments.item_id").
		Where("split_bill_items.payment_id = ? AND split_bill_item_assignments.participant_id = ?", paymentID, participantID).
		Count(&count).Error; err != nil {
		log.WithFields(log.Fields{
			"context":       utils.DumpIncomingContext(ctx),
			"paymentID":     paymentID,
			"participantID": participantID,
		}).Error(err)
		return false, fmt.Errorf("failed to check split bill assignments: %w", err)
	}
	return count > 0, nil
}
