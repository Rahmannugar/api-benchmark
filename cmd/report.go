package cmd

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal"
	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

func printBenchmarkConfiguration(cfg config.BenchmarkConfig) {
	fmt.Printf("Benchmark Target: [%s] %s\n", cfg.Request.Method, cfg.Request.URL)
	fmt.Printf(
		"Requests: %d | Concurrency: %d | Timeout: %v\n",
		cfg.TotalRequests,
		cfg.Concurrency,
		cfg.Request.Timeout,
	)
	if len(cfg.Request.Headers) > 0 {
		fmt.Println("Custom Headers:")
		keys := make([]string, 0, len(cfg.Request.Headers))
		for key := range cfg.Request.Headers {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			fmt.Printf("   - %s: %s\n", key, displayHeaderValue(key, cfg.Request.Headers[key]))
		}
	}
	if len(cfg.Request.Body) > 0 {
		fmt.Printf("Request Body Size: %d bytes\n", len(cfg.Request.Body))
	}
	fmt.Println("Running benchmark, please wait...")
}

// printSummary formats and displays the final benchmark metrics.
func printSummary(s internal.Summary) {
	fmt.Println("\n==================================")
	fmt.Println("BENCHMARK SUMMARY")
	fmt.Println("==================================")
	fmt.Printf("Elapsed Time          : %v\n", s.ElapsedTime)
	fmt.Printf("Requests Attempted    : %d\n", s.AttemptedRequests)
	fmt.Printf("Successful Requests   : %d\n", s.SuccessfulRequests)
	fmt.Printf("Failed Requests       : %d\n", s.FailedRequests)
	fmt.Printf("Estimated Throughput  : %.2f req/sec\n", s.EstimatedThroughput)
	fmt.Printf("Successful Throughput : %.2f req/sec\n", s.SuccessfulThroughput)

	fmt.Println("----------------------------------")
	fmt.Println("SUCCESSFUL REQUEST LATENCY")
	if s.SuccessfulRequests == 0 {
		fmt.Println("No successful requests recorded.")
	} else {
		fmt.Printf("P50                    : %v\n", s.SuccessfulLatency.P50)
		fmt.Printf("P90                    : %v\n", s.SuccessfulLatency.P90)
		fmt.Printf("P95                    : %v\n", s.SuccessfulLatency.P95)
		fmt.Printf("P99                    : %v\n", s.SuccessfulLatency.P99)
		fmt.Printf("Average                : %v\n", s.SuccessfulLatency.Average)
		fmt.Printf("Minimum                : %v\n", s.SuccessfulLatency.Minimum)
		fmt.Printf("Maximum                : %v\n", s.SuccessfulLatency.Maximum)
	}

	if len(s.StatusCodes) > 0 {
		fmt.Println("----------------------------------")
		fmt.Println("STATUS CODE DISTRIBUTION")
		var codes []int
		for code := range s.StatusCodes {
			codes = append(codes, code)
		}
		slices.Sort(codes)
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
		fmt.Println("ERROR BREAKDOWN")
		errorMessages := make([]string, 0, len(s.Errors))
		for errMsg := range s.Errors {
			errorMessages = append(errorMessages, errMsg)
		}
		slices.Sort(errorMessages)
		for _, errMsg := range errorMessages {
			count := s.Errors[errMsg]
			fmt.Printf("   - %s: %d occurrences\n", errMsg, count)
		}
	}
	fmt.Println("==================================")
}

func displayHeaderValue(name, value string) string {
	normalizedName := strings.ToLower(name)
	if normalizedName == "authorization" ||
		normalizedName == "proxy-authorization" ||
		normalizedName == "cookie" ||
		normalizedName == "set-cookie" ||
		strings.Contains(normalizedName, "api-key") ||
		strings.Contains(normalizedName, "token") ||
		strings.Contains(normalizedName, "secret") {
		return "[REDACTED]"
	}

	return value
}
