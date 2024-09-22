package errors

import (
	"errors"
)

var (
	ErrRaceCondition = errors.New("race condition")
)
