package math

import (
	"math"
	"time"
)

func MinDuration(x, y time.Duration) time.Duration {
	if x < math.MinInt64 || y < math.MinInt64 {
		return math.MinInt64
	}
	if x < y {
		return x
	}
	return y
}
