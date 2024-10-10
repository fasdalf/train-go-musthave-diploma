package accrual

import (
	"context"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/catchable"
	"github.com/fasdalf/train-go-musthave-diploma/internal/resources"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strconv"
)

const headerRetryAfter = "Retry-After"

type AccrualClient struct {
	client *resty.Client
	addr   *string
}

func NewAccrualClient(addr *string) *AccrualClient {
	return &AccrualClient{
		client: resty.New(),
		addr:   addr,
	}
}

func (c *AccrualClient) GetAccrual(ctx context.Context, number string) (*resources.AccrualOrderResponse, error) {
	address := fmt.Sprintf("%s/api/orders/%s", *c.addr, number)

	req := c.client.R()
	req.SetContext(ctx)
	req.SetResult(resources.AccrualOrderResponse{})
	resp, err := req.Get(address)
	if err != nil {
		return nil, fmt.Errorf("failed to update accrual order: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("response is empty")
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil, catchable.ErrNotRegistered
	case http.StatusTooManyRequests:
		retryAfterHeader := resp.Header().Get(headerRetryAfter)
		retryAfter, err := strconv.Atoi(retryAfterHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse retry-after header: %w", err)
		}
		return nil, &catchable.ErrorTooManyRequests{Delay: retryAfter}
	case http.StatusOK:
		r, ok := resp.Result().(*resources.AccrualOrderResponse)
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
