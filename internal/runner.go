package internal

import (
	"sync"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

// RunBenchmark executes the configured requests using a bounded worker pool.
func RunBenchmark(cfg config.BenchmarkConfig) []RequestResult {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.TotalRequests <= 0 {
		return nil
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

	// Drain results while workers run to keep the bounded results channel moving.
	allResults := make([]RequestResult, 0, cfg.TotalRequests)
	for res := range results {
		allResults = append(allResults, res)
	}

	return allResults
}
