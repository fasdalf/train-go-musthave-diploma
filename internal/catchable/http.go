package catchable

import (
	"errors"
	"fmt"
)

var (
	ErrNotRegistered       = errors.New("order not registered")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
)

type ErrorTooManyRequests struct {
	Delay int
}

func (e *ErrorTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, Delay %d", e.Delay)
}
