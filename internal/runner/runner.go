package runner

import (
	"context"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/rahmannugar/api-benchmark/internal/config"
	"github.com/rahmannugar/api-benchmark/internal/stats"
)

// RunBenchmark executes the configured requests and returns their aggregate metrics.
func RunBenchmark(ctx context.Context, cfg config.BenchmarkConfig) stats.Summary {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.TotalRequests <= 0 {
		return stats.NewAccumulator(0).Finalize(0)
	}

	client := NewHTTPClient(cfg.Concurrency, cfg.Request.Timeout)
	defer client.CloseIdleConnections()
	pacer := newRequestPacer(cfg.RequestsPerSecond)

	if cfg.WarmupRequests > 0 {
		runRequests(ctx, client, &cfg.Request, cfg.WarmupRequests, cfg.Concurrency, pacer, nil)
		if ctx.Err() != nil {
			return stats.NewAccumulator(0).Finalize(0)
		}
		if !pacer.Prepare(ctx) {
			return stats.NewAccumulator(0).Finalize(0)
		}
	}

	startTime := time.Now()
	accumulator := stats.NewAccumulator(cfg.TotalRequests)
	runRequests(ctx, client, &cfg.Request, cfg.TotalRequests, cfg.Concurrency, pacer, accumulator.Add)

	return accumulator.Finalize(time.Since(startTime))
}

func runRequests(
	ctx context.Context,
	client *http.Client,
	requestConfig *config.RequestConfig,
	totalRequests int,
	concurrency int,
	pacer *requestPacer,
	onResult func(stats.RequestResult),
) {
	workerCount := min(concurrency, totalRequests)
	// Paced jobs use direct handoff so queued work cannot restart in a burst.
	jobBuffer := workerCount
	if pacer.interval > 0 {
		jobBuffer = 0
	}
	jobs := make(chan struct{}, jobBuffer)
	results := make(chan stats.RequestResult, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	// Each worker processes one request at a time and then takes the next job.
	for range workerCount {
		go func() {
			defer wg.Done()
			for range jobs {
				if ctx.Err() != nil {
					return
				}
				results <- SendRequest(ctx, client, requestConfig)
			}
		}()
	}

	// Produce jobs independently so workers can begin before every job is queued.
	go produceJobs(ctx, jobs, totalRequests, pacer)

	// Results are complete only after every worker has stopped sending.
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if onResult != nil {
			onResult(result)
		}
	}
}

func produceJobs(
	ctx context.Context,
	jobs chan<- struct{},
	totalRequests int,
	pacer *requestPacer,
) {
	defer close(jobs)
	for range totalRequests {
		if !pacer.Wait(ctx) {
			return
		}
		select {
		case jobs <- struct{}{}:
		case <-ctx.Done():
			return
		}
	}
}

type requestPacer struct {
	interval  time.Duration
	nextStart time.Time
	prepared  bool
}

func newRequestPacer(requestsPerSecond float64) *requestPacer {
	if requestsPerSecond == 0 {
		return &requestPacer{}
	}

	intervalNanoseconds := float64(time.Second) / requestsPerSecond
	intervalNanoseconds = max(1, min(intervalNanoseconds, float64(math.MaxInt64)))
	return &requestPacer{interval: time.Duration(intervalNanoseconds)}
}

func (p *requestPacer) Wait(ctx context.Context) bool {
	if p.prepared {
		p.prepared = false
		return true
	}
	if p.interval == 0 {
		return true
	}

	delay := time.Until(p.nextStart)
	if delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return false
		}
	}

	p.nextStart = time.Now().Add(p.interval)
	return true
}

func (p *requestPacer) Prepare(ctx context.Context) bool {
	if !p.Wait(ctx) {
		return false
	}
	p.prepared = true
	return true
}
