package internal

import (
	"sync"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

// RunBenchmark executes the configured requests and returns their aggregate metrics.
func RunBenchmark(cfg config.BenchmarkConfig) Summary {
	startTime := time.Now()
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.TotalRequests <= 0 {
		return newStatsAccumulator(0).finalize(time.Since(startTime))
	}

	workerCount := min(cfg.Concurrency, cfg.TotalRequests)
	client := NewHTTPClient(workerCount, cfg.Request.Timeout)

	// Bound channel memory by worker count rather than total request count.
	jobs := make(chan struct{}, workerCount)
	results := make(chan RequestResult, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	// Each worker processes one request at a time and then takes the next job.
	for range workerCount {
		go func() {
			defer wg.Done()
			for range jobs {
				results <- SendRequest(client, &cfg.Request)
			}
		}()
	}

	// Produce jobs independently so workers can begin before every job is queued.
	go func() {
		defer close(jobs)
		for range cfg.TotalRequests {
			jobs <- struct{}{}
		}
	}()

	// Results are complete only after every worker has stopped sending.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate each result immediately instead of retaining every full result.
	stats := newStatsAccumulator(workerCount)
	for result := range results {
		stats.add(result)
	}

	return stats.finalize(time.Since(startTime))
}
