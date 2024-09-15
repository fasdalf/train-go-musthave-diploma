package accrual

import (
	"context"
	"errors"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	resty "github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	workersCount  = 3
	fetchTimeout  = 300 * time.Millisecond
	fetchInterval = 400 * time.Millisecond
	fetchFactor   = 10

	orderStatusRegistered = "REGISTERED"
	orderStatusInvalid    = "INVALID"
	orderStatusProcessing = "PROCESSING"
	orderStatusProcessed  = "PROCESSED"

	headerRetryAfter = "Retry-After"
)

type accrualOrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type errorTooManyRequests struct {
	delay int
}

func (e *errorTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, delay %d", e.delay)
}

var errorNotRegistered = fmt.Errorf("order not registered")

func StartChecker(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB, addr string) {
	wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go worker(ctx, wg, db, addr, i+1)
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB, addr string, id int) {
	defer wg.Done()
	idlog := slog.With("worker", "accrualChecker", "id", id)
	client := resty.New()
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		timer.Stop()
		err := (func() error {
			c2, cancel2 := context.WithTimeout(ctx, fetchTimeout)
			defer cancel2()

			order, err := getOrderFromDB(c2, db)
			if err != nil {
				return err
			}

			accrual, err := getAccrual(c2, client, addr, order.OrderNumber)
			if err != nil {
				_ = setOrderStatus(db, order, getOrderFetchStatus(accrual, err))
				return err
			}

			err = saveCheckResult(c2, db, order, accrual)
			if err != nil {
				return err
			}

			return nil
		})()

		delay := time.Duration(0)
		if err != nil {
			idlog.Error("error while processing", "error", err)
			// Sleep for some time if no new jobs.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				delay = fetchInterval
			}
			// Sleep for retry-after seconds
			var errDelay *errorTooManyRequests
			if errors.As(err, &errDelay) && errDelay != nil {
				delay = time.Duration(errDelay.delay)
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

func getOrderFromDB(ctx context.Context, db *gorm.DB) (*entity.Order, error) {
	var order *entity.Order
	threshold := time.Now().Add(-fetchTimeout * fetchFactor)

	tx := db.WithContext(ctx).Begin()

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
		return nil, fmt.Errorf("race detected for order %s", order.OrderNumber)
	}

	if err := setOrderStatus(tx, order, entity.FetchStatusInProgress); err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return order, nil
}

func getOrderFetchStatus(accrual *accrualOrderResponse, err error) string {
	if errors.Is(err, errorNotRegistered) {
		return entity.FetchStatusFailure
	}
	if accrual == nil {
		return entity.FetchStatusWaiting
	}
	return entity.FetchStatusInProgress
}

func setOrderStatus(db *gorm.DB, order *entity.Order, status string) error {
	order.FetchStatus = status
	db.Save(order)
	return db.Error
}

func getAccrual(ctx context.Context, client *resty.Client, addr string, number string) (*accrualOrderResponse, error) {
	address := fmt.Sprintf("%s/api/orders/%s", addr, number)

	req := client.R()
	req.SetContext(ctx)
	req.SetResult(accrualOrderResponse{})
	resp, err := req.Get(address)
	if err != nil {
		return nil, fmt.Errorf("failed to update accrual order: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("response is empty")
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil, errorNotRegistered
	case http.StatusTooManyRequests:
		retryAfterHeader := resp.Header().Get(headerRetryAfter)
		retryAfter, err := strconv.Atoi(retryAfterHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse retry-after header: %d", resp.StatusCode())
		}
		return nil, &errorTooManyRequests{delay: retryAfter}
	case http.StatusOK:
		r, ok := resp.Result().(*accrualOrderResponse)
		if !ok {
			return nil, fmt.Errorf("response is not a")
		}

		if number != r.Order {
			return nil, fmt.Errorf("order number mismatch")
		}

		return r, nil
	}
	return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode())
}

func saveCheckResult(ctx context.Context, db *gorm.DB, order *entity.Order, accrual *accrualOrderResponse) error {
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

	tx := db.WithContext(ctx).Begin()

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
