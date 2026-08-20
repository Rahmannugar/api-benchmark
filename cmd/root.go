package cmd

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal"
	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

// headerFlags is a custom flag type to allow multiple -H / -header arguments
type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

// Execute parses CLI flags and runs the benchmark process
func Execute() error {
	// Define command-line flags
	urlFlag := flag.String("url", "http://localhost:8080", "Target URL to benchmark")
	methodFlag := flag.String("m", "GET", "HTTP method (GET, POST, PUT, DELETE, PATCH, etc.)")
	flag.StringVar(methodFlag, "method", "GET", "HTTP method (alias for -m)")

	requestsFlag := flag.Int("n", 1000, "Total number of requests to perform")
	concurrencyFlag := flag.Int("c", 10, "Number of concurrent workers")

	dataFlag := flag.String("d", "", "HTTP request body / data")
	flag.StringVar(dataFlag, "data", "", "HTTP request body / data (alias for -d)")
	flag.StringVar(dataFlag, "body", "", "HTTP request body / data (alias for -d)")

	var headers headerFlags
	flag.Var(&headers, "H", "Custom HTTP header (e.g. -H \"Content-Type: application/json\"), repeatable")
	flag.Var(&headers, "header", "Custom HTTP header (alias for -H), repeatable")

	timeoutFlag := flag.Duration("t", 10*time.Second, "Per-request timeout (e.g. 5s, 10s)")
	flag.DurationVar(timeoutFlag, "timeout", 10*time.Second, "Per-request timeout (alias for -t)")

	flag.Parse()

	method := strings.ToUpper(strings.TrimSpace(*methodFlag))
	if method == "" {
		method = "GET"
	}

	// Parse headers into map
	headerMap := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else {
			headerMap[strings.TrimSpace(parts[0])] = ""
		}
	}

	reqConfig := config.RequestConfig{
		Method:  method,
		URL:     *urlFlag,
		Headers: headerMap,
		Body:    []byte(*dataFlag),
		Timeout: *timeoutFlag,
	}

	benchConfig := config.BenchmarkConfig{
		Request:       reqConfig,
		TotalRequests: *requestsFlag,
		Concurrency:   *concurrencyFlag,
	}
	if err := benchConfig.Validate(); err != nil {
		return fmt.Errorf("invalid benchmark configuration: %w", err)
	}

	fmt.Printf("🎯 Benchmarking Target: [%s] %s\n", method, *urlFlag)
	fmt.Printf("📦 Total Requests: %d | Concurrency: %d | Timeout: %v\n", *requestsFlag, *concurrencyFlag, *timeoutFlag)
	if len(headerMap) > 0 {
		fmt.Println("📋 Custom Headers:")
		for k, v := range headerMap {
			fmt.Printf("   - %s: %s\n", k, v)
		}
	}
	if len(*dataFlag) > 0 {
		fmt.Printf("📝 Request Body Size: %d bytes\n", len(*dataFlag))
	}
	fmt.Println("⏳ Running benchmark, please wait...")

	startTime := time.Now()
	results := internal.RunBenchmark(benchConfig)
	totalDuration := time.Since(startTime)

	// Calculate and print the performance summary
	summary := internal.CalculateStats(results, totalDuration)
	printSummary(summary)

	return nil
}
