package utils

import (
	"time"
)

func IsTimeEqualWithHysteresis(t1 time.Time, t2 time.Time, hyst time.Duration) bool {
	diff := t1.Sub(t2).Abs()
	return diff <= hyst
}
