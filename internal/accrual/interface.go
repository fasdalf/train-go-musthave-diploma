package accrual

import (
	"context"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	"github.com/fasdalf/train-go-musthave-diploma/internal/resources"
	"time"
)

type OrderRepository interface {
	GetOrderByThreshold(ctx context.Context, threshold time.Time) (*entity.Order, error)
	SaveOrder(order *entity.Order) error
	SaveOrderWithUserAccrual(ctx context.Context, order *entity.Order) error
}

type AccrualClient interface {
	GetAccrual(ctx context.Context, number string) (*resources.AccrualOrderResponse, error)
}
