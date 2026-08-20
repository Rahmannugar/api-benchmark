package internal

import (
	"math"
	"slices"
	"time"
)

// CalculateStats processes the list of results and computes performance metrics
func CalculateStats(results []RequestResult, elapsedTime time.Duration) Summary {
	summary := Summary{
		AttemptedRequests: len(results),
		ElapsedTime:       elapsedTime,
		StatusCodes:       make(map[int]int),
		Errors:            make(map[string]int),
	}
	if len(results) == 0 {
		return summary
	}

	successfulDurations := make([]time.Duration, 0, len(results))

	for _, res := range results {
		if res.StatusCode > 0 {
			summary.StatusCodes[res.StatusCode]++
		}

		if res.Error != nil {
			summary.Errors[res.Error.Error()]++
		}

		if !isSuccessfulResult(res) {
			summary.FailedRequests++
			continue
		}

		summary.SuccessfulRequests++
		successfulDurations = append(successfulDurations, res.Duration)
	}

	summary.SuccessfulLatency = calculateLatencyStats(successfulDurations)

	if elapsedTime.Seconds() > 0 {
		summary.EstimatedThroughput = float64(summary.AttemptedRequests) / elapsedTime.Seconds()
		summary.SuccessfulThroughput = float64(summary.SuccessfulRequests) / elapsedTime.Seconds()
	}

	return summary
}

func isSuccessfulResult(result RequestResult) bool {
	return result.Error == nil && result.StatusCode >= 200 && result.StatusCode < 400
}

func calculateLatencyStats(durations []time.Duration) LatencyStats {
	if len(durations) == 0 {
		return LatencyStats{}
	}

	slices.Sort(durations)

	var total time.Duration
	for _, duration := range durations {
		total += duration
	}

	return LatencyStats{
		Average: total / time.Duration(len(durations)),
		Minimum: durations[0],
		Maximum: durations[len(durations)-1],
		P50:     percentile(durations, 0.50),
		P90:     percentile(durations, 0.90),
		P95:     percentile(durations, 0.95),
		P99:     percentile(durations, 0.99),
	}
}

func percentile(sortedDurations []time.Duration, percentile float64) time.Duration {
	index := int(math.Ceil(percentile*float64(len(sortedDurations)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sortedDurations) {
		index = len(sortedDurations) - 1
	}

	return sortedDurations[index]
}
