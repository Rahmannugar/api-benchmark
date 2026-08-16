package internal

import (
	"time"
)
type Summary struct {
	TotalRequests int
	SuccessfulRequests int
	FailedRequests int
	TotalDuration time.Duration
	RequestPerSec float64
	MinLatency time.Duration
	MaxLatency time.Duration
	AvgLatency time.Duration

}
// CalculateStats processes the list of results and computes performance metrics
func CalculateStats(results []Result, totalDuration time.Duration) Summary{
	summary := Summary{
		TotalRequests: len(results),
		TotalDuration: totalDuration,
		
	}
	if len(results)==0{
		return summary
	}
	
	var totalLatency time.Duration
	summary.MinLatency=results[0].Duration
	summary.MaxLatency=results[0].Duration

	for _,res:=range results {
		// Count successful vs failed requests based on errors or HTTP status codes
		if res.Error != nil || res.StatusCode >= 400 {
			summary.FailedRequests++
		}else {
			summary.SuccessfulRequests++
		}

		totalLatency += res.Duration

		// Track minimum latency
		if res.Duration < summary.MinLatency {
			summary.MinLatency = res.Duration
		}
		// Track maximum latency
		if res.Duration > summary.MaxLatency {
			summary.MaxLatency = res.Duration
		}
	}

	// Calculate average latency
	summary.AvgLatency = totalLatency / time.Duration(len(results))

	// Calculate Requests Per Second (RPS)
	if totalDuration.Seconds() > 0 {
		summary.RequestPerSec = float64(summary.TotalRequests) / totalDuration.Seconds()
	}

	return summary
}