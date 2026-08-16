package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal"
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
func Execute() {
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

	reqConfig := internal.RequestConfig{
		Method:  method,
		URL:     *urlFlag,
		Headers: headerMap,
		Body:    []byte(*dataFlag),
		Timeout: *timeoutFlag,
	}

	benchConfig := internal.BenchmarkConfig{
		Request:       reqConfig,
		TotalRequests: *requestsFlag,
		Concurrency:   *concurrencyFlag,
	}

	startTime := time.Now()
	results := internal.RunBenchmark(benchConfig)
	totalDuration := time.Since(startTime)

	// Calculate and print the performance summary
	summary := internal.CalculateStats(results, totalDuration)
	printSummary(summary)
}

// printSummary formats and displays the final benchmark metrics
func printSummary(s internal.Summary) {
	fmt.Println("\n==================================")
	fmt.Println("📊 BENCHMARK RESULTS SUMMARY")
	fmt.Println("==================================")
	fmt.Printf("Total Time Taken   : %v\n", s.TotalDuration)
	fmt.Printf("Successful Requests: %d\n", s.SuccessfulRequests)
	fmt.Printf("Failed Requests    : %d\n", s.FailedRequests)
	fmt.Printf("Requests Per Sec   : %.2f req/sec\n", s.RequestPerSec)
	fmt.Printf("Average Latency    : %v\n", s.AvgLatency)
	fmt.Printf("Minimum Latency    : %v\n", s.MinLatency)
	fmt.Printf("Maximum Latency    : %v\n", s.MaxLatency)

	if len(s.StatusCodes) > 0 {
		fmt.Println("----------------------------------")
		fmt.Println("📈 Status Codes Distribution:")
		var codes []int
		for code := range s.StatusCodes {
			codes = append(codes, code)
		}
		sort.Ints(codes)
		for _, code := range codes {
			statusText := http.StatusText(code)
			if statusText == "" {
				statusText = "Unknown"
			}
			fmt.Printf("   [%d %s]: %d responses\n", code, statusText, s.StatusCodes[code])
		}
	}

	if len(s.Errors) > 0 {
		fmt.Println("----------------------------------")
		fmt.Println("❌ Error Breakdown:")
		for errMsg, count := range s.Errors {
			fmt.Printf("   - %s: %d occurrences\n", errMsg, count)
		}
	}
	fmt.Println("==================================")
}