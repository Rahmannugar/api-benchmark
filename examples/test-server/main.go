package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	serverAddress     = "localhost:8080"
	coldStartRequests = 3
)

func main() {
	var benchmarkRequestCount atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "healthy")
	})
	mux.HandleFunc("GET /benchmark", benchmarkHandler(&benchmarkRequestCount))

	server := &http.Server{
		Addr:              serverAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("example server listening on http://%s", serverAddress)
	log.Printf("health endpoint: http://%s/health", serverAddress)
	log.Printf("benchmark endpoint: http://%s/benchmark", serverAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func benchmarkHandler(requestCount *atomic.Int64) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		requestNumber := requestCount.Add(1)
		statusCode := http.StatusOK
		delay := 20 * time.Millisecond

		if requestNumber <= coldStartRequests {
			delay = 750 * time.Millisecond
		} else if (requestNumber-coldStartRequests)%4 == 0 {
			statusCode = http.StatusInternalServerError
			delay = 100 * time.Millisecond
		}

		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = fmt.Fprintf(
			w,
			`{"request":%d,"status":%d,"simulated_delay_ms":%d}`+"\n",
			requestNumber,
			statusCode,
			delay.Milliseconds(),
		)
		log.Printf("request=%d status=%d delay=%s", requestNumber, statusCode, delay)
	}
}
