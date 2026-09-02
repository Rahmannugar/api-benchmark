package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rahmannugar/api-benchmark/cmd/report"
	"github.com/rahmannugar/api-benchmark/internal/config"
	"github.com/rahmannugar/api-benchmark/internal/runner"
)

// headerFlags collects repeated header arguments.
// Use -H for headers; lowercase -h displays CLI help.
type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

// Execute parses CLI flags and runs the benchmark process.
func Execute(ctx context.Context) error {
	// targetURL is the endpoint to benchmark.
	var targetURL string
	flag.StringVar(&targetURL, "url", "http://localhost:8080", "Target URL to benchmark")

	// method is the HTTP method used for each request.
	var method string
	flag.StringVar(&method, "m", "GET", "HTTP method")
	flag.StringVar(&method, "method", "GET", "HTTP method (alias for -m)")

	// totalRequests is the number of requests included in the results.
	var totalRequests int
	flag.IntVar(&totalRequests, "n", 1000, "Total number of measured requests")
	flag.IntVar(&totalRequests, "requests", 1000, "Total measured requests (alias for -n)")

	// warmupRequests is the number of requests sent before measurement begins.
	var warmupRequests int
	flag.IntVar(&warmupRequests, "w", 0, "Warm-up requests performed before measurement")
	flag.IntVar(&warmupRequests, "warmup", 0, "Warm-up requests (alias for -w)")
	flag.IntVar(&warmupRequests, "warm-up", 0, "Warm-up requests (alias for -w)")

	// concurrency is the maximum number of requests that may run at once.
	var concurrency int
	flag.IntVar(&concurrency, "c", 10, "Number of concurrent workers")
	flag.IntVar(&concurrency, "concurrency", 10, "Concurrent workers (alias for -c)")

	// requestsPerSecond limits how many requests may start each second.
	var requestsPerSecond float64
	flag.Float64Var(&requestsPerSecond, "r", 0, "Maximum requests started per second; 0 means unlimited")
	flag.Float64Var(&requestsPerSecond, "rate", 0, "Maximum request start rate (alias for -r)")

	// headers contains every custom header provided with -H or --header.
	var headers headerFlags
	flag.Var(&headers, "H", "Custom HTTP header (e.g. -H \"Content-Type: application/json\"), repeatable")
	flag.Var(&headers, "header", "Custom HTTP header (alias for -H), repeatable")

	// requestBody is the optional body sent with each request.
	var requestBody string
	flag.StringVar(&requestBody, "d", "", "HTTP request body / data")
	flag.StringVar(&requestBody, "data", "", "HTTP request body / data (alias for -d)")
	flag.StringVar(&requestBody, "body", "", "HTTP request body / data (alias for -d)")

	// requestTimeout is the maximum time allowed for one request.
	var requestTimeout time.Duration
	flag.DurationVar(&requestTimeout, "t", 10*time.Second, "Per-request timeout (e.g. 5s, 10s)")
	flag.DurationVar(&requestTimeout, "timeout", 10*time.Second, "Per-request timeout (alias for -t)")

	// outputFormat selects text or JSON results.
	var outputFormat string
	flag.StringVar(&outputFormat, "format", "text", "Output format: text or json")

	flag.Parse()

	method = strings.ToUpper(strings.TrimSpace(method))
	outputFormat = strings.ToLower(strings.TrimSpace(outputFormat))
	if outputFormat != "text" && outputFormat != "json" {
		return fmt.Errorf("invalid output format %q: expected text or json", outputFormat)
	}

	headerMap, err := config.ParseHeaders([]string(headers))
	if err != nil {
		return err
	}

	reqConfig := config.RequestConfig{
		Method:  method,
		URL:     targetURL,
		Headers: headerMap,
		Body:    []byte(requestBody),
		Timeout: requestTimeout,
	}

	benchConfig := config.BenchmarkConfig{
		Request:           reqConfig,
		TotalRequests:     totalRequests,
		WarmupRequests:    warmupRequests,
		Concurrency:       concurrency,
		RequestsPerSecond: requestsPerSecond,
	}
	if err := benchConfig.Validate(); err != nil {
		return fmt.Errorf("invalid benchmark configuration: %w", err)
	}

	if outputFormat == "text" {
		report.PrintBenchmarkConfiguration(benchConfig)
	}

	summary := runner.RunBenchmark(ctx, benchConfig)
	if outputFormat == "json" {
		if err := report.WriteJSON(os.Stdout, benchConfig, summary); err != nil {
			return fmt.Errorf("write JSON report: %w", err)
		}
	} else {
		report.PrintSummary(summary)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("benchmark interrupted: %w", err)
	}

	return nil
}
