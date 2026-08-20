package cmd

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal"
)

// printSummary formats and displays the final benchmark metrics.
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
