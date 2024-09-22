package repository

import (
	"context"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	internalErrors "github.com/fasdalf/train-go-musthave-diploma/internal/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetOrderByThreshold(ctx context.Context, threshold time.Time) (*entity.Order, error) {
	var order *entity.Order

	tx := r.db.WithContext(ctx).Begin()

	query := tx.Model(&entity.Order{}).Where(
		"fetch_status = ? OR (fetch_status = ? AND updated_at < ?)",
		entity.FetchStatusWaiting,
		entity.FetchStatusInProgress,
		threshold,
	).Order("updated_at").Limit(1).Clauses(clause.Locking{Strength: "UPDATE"})
	if err := query.Find(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if order.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	if order.FetchStatus != entity.FetchStatusWaiting && order.UpdatedAt.After(threshold) {
		return nil, fmt.Errorf("%w for order %s", internalErrors.ErrRaceCondition, order.OrderNumber)
	}

	order.FetchStatus = entity.FetchStatusInProgress
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) SaveOrderWithUserAccrual(ctx context.Context, order *entity.Order) error {
	tx := r.db.WithContext(ctx).Begin()

	if order.Amount > 0 {
		if err := tx.Exec("UPDATE users SET amount = amount + ? WHERE id = ?", order.Amount, order.UserID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *OrderRepository) SaveOrder(order *entity.Order) error {
	return r.db.Save(order).Error
}
