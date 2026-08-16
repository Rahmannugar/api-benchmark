package cmd

import (
	"flag"
	"fmt"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal"
)

// Execute parses CLI flags and runs the benchmark process
func Execute() {
	// Define command-line flags
	urlFlag := flag.String("url", "http://localhost:8080", "Target URL to benchmark")
	requestsFlag := flag.Int("n", 1000, "Total number of requests to perform")
	concurrencyFlag := flag.Int("c", 10, "Number of concurrent workers")

	flag.Parse()

	fmt.Printf("🎯 Benchmarking Target: %s\n", *urlFlag)
	fmt.Printf("📦 Total Requests: %d | Concurrency: %d\n", *requestsFlag, *concurrencyFlag)
	fmt.Println("⏳ Running benchmark, please wait...")

	startTime := time.Now()
	results := internal.RunBenchmark(*urlFlag, *requestsFlag, *concurrencyFlag)
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
	fmt.Println("==================================")
}