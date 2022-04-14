package constants

import "time"

var (
	// retry function with 10 times, delay 100ms
	RetryAttempts = uint(10)
	RetryDelay    = 100 * time.Millisecond
)
