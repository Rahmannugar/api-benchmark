package internal

import "time"

// Result holds the metrics for a single HTTP request execution.
type Result struct {
	Duration   time.Duration
	Error      error
	StatusCode int
}
