package repository

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"

	"gorm.io/gorm"
)

type (
	gormTransactioner struct {
		db *gorm.DB
	}
)

// NewGormTransactioner create new instance
func NewGormTransactioner(db *gorm.DB) model.GormTransactioner {
	return &gormTransactioner{db: db}
}

// Begin transaction
func (t *gormTransactioner) Begin(ctx context.Context) *gorm.DB {
	return t.db.WithContext(ctx).Begin()
}

// Commit transaction
func (t *gormTransactioner) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback transaction
func (t *gormTransactioner) Rollback(tx *gorm.DB) {
	tx.Rollback()
}
