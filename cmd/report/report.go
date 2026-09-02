package report

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/rahmannugar/api-benchmark/internal/config"
	"github.com/rahmannugar/api-benchmark/internal/stats"
)

type jsonReport struct {
	Target        jsonTarget        `json:"target"`
	Configuration jsonConfiguration `json:"configuration"`
	Summary       jsonSummary       `json:"summary"`
}

type jsonTarget struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type jsonConfiguration struct {
	TotalRequests        int               `json:"total_requests"`
	WarmupRequests       int               `json:"warmup_requests"`
	Concurrency          int               `json:"concurrency"`
	MaxRequestsPerSecond float64           `json:"max_requests_per_second"`
	TimeoutSeconds       float64           `json:"timeout_seconds"`
	Headers              map[string]string `json:"headers"`
	BodySizeBytes        int               `json:"body_size_bytes"`
}

type jsonSummary struct {
	ElapsedSeconds      float64        `json:"elapsed_seconds"`
	AttemptedRequests   int            `json:"attempted_requests"`
	SuccessfulRequests  int            `json:"successful_requests"`
	FailedRequests      int            `json:"failed_requests"`
	Throughput          jsonThroughput `json:"throughput"`
	SuccessfulLatencyMS jsonLatency    `json:"successful_latency_ms"`
	FailedLatencyMS     jsonLatency    `json:"failed_latency_ms"`
	AllLatencyMS        jsonLatency    `json:"all_requests_latency_ms"`
	StatusCodes         map[int]int    `json:"status_codes"`
	Errors              map[string]int `json:"errors"`
}

type jsonThroughput struct {
	EstimatedRequestsPerSecond  float64 `json:"estimated_requests_per_second"`
	SuccessfulRequestsPerSecond float64 `json:"successful_requests_per_second"`
}

type jsonLatency struct {
	Average float64 `json:"average"`
	Minimum float64 `json:"minimum"`
	Maximum float64 `json:"maximum"`
	P50     float64 `json:"p50"`
	P90     float64 `json:"p90"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
}

// PrintBenchmarkConfiguration displays the validated text-mode configuration.
func PrintBenchmarkConfiguration(cfg config.BenchmarkConfig) {
	fmt.Printf("Benchmark Target: [%s] %s\n", cfg.Request.Method, cfg.Request.URL)
	fmt.Printf(
		"Requests: %d | Warm-up: %d | Concurrency: %d | Rate: %s | Timeout: %v\n",
		cfg.TotalRequests,
		cfg.WarmupRequests,
		cfg.Concurrency,
		formatRequestRate(cfg.RequestsPerSecond),
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

// PrintSummary formats and displays the final text benchmark metrics.
func PrintSummary(s stats.Summary) {
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
	printLatency("FAILED REQUEST LATENCY", s.FailedRequests, "failed", s.FailedLatency)
	printLatency("ALL REQUEST LATENCY", s.AttemptedRequests, "measured", s.AllLatency)

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

func printLatency(title string, requestCount int, requestType string, latency stats.LatencyStats) {
	fmt.Println("----------------------------------")
	fmt.Println(title)
	if requestCount == 0 {
		fmt.Printf("No %s requests recorded.\n", requestType)
		return
	}
	fmt.Printf("P50                    : %v\n", latency.P50)
	fmt.Printf("P90                    : %v\n", latency.P90)
	fmt.Printf("P95                    : %v\n", latency.P95)
	fmt.Printf("P99                    : %v\n", latency.P99)
	fmt.Printf("Average                : %v\n", latency.Average)
	fmt.Printf("Minimum                : %v\n", latency.Minimum)
	fmt.Printf("Maximum                : %v\n", latency.Maximum)
}

// WriteJSON writes one machine-readable benchmark report.
func WriteJSON(writer io.Writer, cfg config.BenchmarkConfig, summary stats.Summary) error {
	report := jsonReport{
		Target: jsonTarget{
			Method: cfg.Request.Method,
			URL:    cfg.Request.URL,
		},
		Configuration: jsonConfiguration{
			TotalRequests:        cfg.TotalRequests,
			WarmupRequests:       cfg.WarmupRequests,
			Concurrency:          cfg.Concurrency,
			MaxRequestsPerSecond: cfg.RequestsPerSecond,
			TimeoutSeconds:       cfg.Request.Timeout.Seconds(),
			Headers:              displayHeaders(cfg.Request.Headers),
			BodySizeBytes:        len(cfg.Request.Body),
		},
		Summary: jsonSummary{
			ElapsedSeconds:     summary.ElapsedTime.Seconds(),
			AttemptedRequests:  summary.AttemptedRequests,
			SuccessfulRequests: summary.SuccessfulRequests,
			FailedRequests:     summary.FailedRequests,
			Throughput: jsonThroughput{
				EstimatedRequestsPerSecond:  summary.EstimatedThroughput,
				SuccessfulRequestsPerSecond: summary.SuccessfulThroughput,
			},
			SuccessfulLatencyMS: jsonLatency{
				Average: durationMilliseconds(summary.SuccessfulLatency.Average),
				Minimum: durationMilliseconds(summary.SuccessfulLatency.Minimum),
				Maximum: durationMilliseconds(summary.SuccessfulLatency.Maximum),
				P50:     durationMilliseconds(summary.SuccessfulLatency.P50),
				P90:     durationMilliseconds(summary.SuccessfulLatency.P90),
				P95:     durationMilliseconds(summary.SuccessfulLatency.P95),
				P99:     durationMilliseconds(summary.SuccessfulLatency.P99),
			},
			FailedLatencyMS: toJSONLatency(summary.FailedLatency),
			AllLatencyMS:    toJSONLatency(summary.AllLatency),
			StatusCodes:     summary.StatusCodes,
			Errors:          summary.Errors,
		},
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func toJSONLatency(latency stats.LatencyStats) jsonLatency {
	return jsonLatency{
		Average: durationMilliseconds(latency.Average),
		Minimum: durationMilliseconds(latency.Minimum),
		Maximum: durationMilliseconds(latency.Maximum),
		P50:     durationMilliseconds(latency.P50),
		P90:     durationMilliseconds(latency.P90),
		P95:     durationMilliseconds(latency.P95),
		P99:     durationMilliseconds(latency.P99),
	}
}

func displayHeaders(headers map[string]string) map[string]string {
	displayed := make(map[string]string, len(headers))
	for name, value := range headers {
		displayed[name] = displayHeaderValue(name, value)
	}

	return displayed
}

func durationMilliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}

func formatRequestRate(requestsPerSecond float64) string {
	if requestsPerSecond == 0 {
		return "unlimited"
	}

	return fmt.Sprintf("%g req/sec", requestsPerSecond)
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
