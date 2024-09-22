package catchable

import (
	"errors"
)

var (
	ErrRaceCondition = errors.New("race condition")
)
