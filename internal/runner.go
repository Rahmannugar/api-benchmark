package internal

import (
	"sync"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

// RunBenchmark executes totalRequests concurrently using the specified number of workers
func RunBenchmark(cfg config.BenchmarkConfig) []RequestResult {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.TotalRequests <= 0 {
		return nil
	}

	client := NewHTTPClient(cfg.Concurrency, cfg.Request.Timeout)

	jobs := make(chan struct{}, cfg.TotalRequests)
	results := make(chan RequestResult, cfg.TotalRequests)

	// Feed all jobs into the channel
	for i := 0; i < cfg.TotalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var wg sync.WaitGroup

	// Spawn worker goroutines
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				res := SendRequest(client, &cfg.Request)
				results <- res
			}
		}()
	}

	wg.Wait()
	close(results)

	// Collect all results from the channel into a slice
	allResults := make([]RequestResult, 0, cfg.TotalRequests)
	for res := range results {
		allResults = append(allResults, res)
	}

	return allResults
}
