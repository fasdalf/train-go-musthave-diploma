package accrual

import (
	"context"
	"errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	internalErrors "github.com/fasdalf/train-go-musthave-diploma/internal/errors"
	"github.com/fasdalf/train-go-musthave-diploma/internal/resources"
	"gorm.io/gorm"
	"log/slog"
	"sync"
	"time"
)

const (
	orderStatusRegistered = "REGISTERED"
	orderStatusInvalid    = "INVALID"
	orderStatusProcessing = "PROCESSING"
	orderStatusProcessed  = "PROCESSED"
)

type Config struct {
	URL           string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	WorkersCount  int           `env:"ACCRUAL_WORKERS_COUNT"`
	FetchTimeout  time.Duration `env:"ACCRUAL_FETCH_TIMEOUT"`
	FetchInterval time.Duration `env:"ACCRUAL_FETCH_INTERVAL"`
	FetchFactor   int           `env:"ACCRUAL_FETCH_FACTOR"`
}

func StartChecker(ctx context.Context, wg *sync.WaitGroup, ac AccrualClient, r OrderRepository, cfg *Config) {
	wg.Add(cfg.WorkersCount)
	for i := 0; i < cfg.WorkersCount; i++ {
		go worker(ctx, wg, ac, r, cfg, i+1)
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, ac AccrualClient, r OrderRepository, cfg *Config, id int) {
	defer wg.Done()
	idlog := slog.With("worker", "accrualChecker", "id", id)
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		timer.Stop()
		delay := time.Duration(0)
		err := processOneOrder(ctx, cfg, ac, r)
		if err != nil {
			idlog.Error("error while processing", "error", err)
			// Sleep for some time if no new jobs.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				delay = cfg.FetchInterval
			}
			// Sleep for retry-after seconds
			var errDelay *internalErrors.ErrorTooManyRequests
			if errors.As(err, &errDelay) && errDelay != nil {
				delay = time.Duration(errDelay.Delay) * time.Second
			}
		}
		timer.Reset(delay)

		select {
		case <-ctx.Done():
			idlog.Info("context canceled, stopping worker")
			return
		case <-timer.C:
		}
	}
}

func processOneOrder(ctx context.Context, cfg *Config, ac AccrualClient, r OrderRepository) error {
	c2, cancel2 := context.WithTimeout(ctx, cfg.FetchTimeout)
	defer cancel2()

	threshold := time.Now().Add(-cfg.FetchTimeout * time.Duration(cfg.FetchFactor))
	order, err := r.GetOrderByThreshold(c2, threshold)
	if err != nil {
		return err
	}

	accrual, err := ac.GetAccrual(c2, order.OrderNumber)
	if err != nil {
		order.FetchStatus = getOrderFetchStatus(accrual, err)
		_ = r.SaveOrder(order)
		return err
	}

	err = saveCheckResult(c2, r, order, accrual)
	if err != nil {
		return err
	}

	return nil
}

func getOrderFetchStatus(accrual *resources.AccrualOrderResponse, err error) string {
	if errors.Is(err, internalErrors.ErrNotRegistered) {
		return entity.FetchStatusFailure
	}
	if accrual == nil {
		return entity.FetchStatusWaiting
	}
	return entity.FetchStatusInProgress
}

func saveCheckResult(ctx context.Context, r OrderRepository, order *entity.Order, accrual *resources.AccrualOrderResponse) error {
	order.OrderStatus = accrual.Status

	switch accrual.Status {
	case orderStatusRegistered:
	case orderStatusProcessing:
		order.FetchStatus = entity.FetchStatusWaiting
	case orderStatusInvalid:
		order.FetchStatus = entity.FetchStatusFailure
	case orderStatusProcessed:
		order.FetchStatus = entity.FetchStatusFinished
		order.Amount = accrual.Accrual
	}

	return r.SaveOrderWithUserAccrual(ctx, order)
}
