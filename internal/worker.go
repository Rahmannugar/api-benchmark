package internal

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

// Result holds the metrics for a single HTTP request execution
type Result struct {
	Duration   time.Duration
	Error      error
	StatusCode int
}

// NewHTTPClient creates an optimized HTTP client with connection pooling scaled for concurrency
func NewHTTPClient(concurrency int, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        concurrency * 2,
		MaxIdleConnsPerHost: concurrency * 2,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// SendRequest executes a single HTTP request with the provided configuration and measures its duration
func SendRequest(client *http.Client, requestConfig *config.RequestConfig) Result {
	var bodyReader io.Reader
	if len(requestConfig.Body) > 0 {
		bodyReader = bytes.NewReader(requestConfig.Body)
	}

	req, err := http.NewRequest(requestConfig.Method, requestConfig.URL, bodyReader)
	if err != nil {
		return Result{
			Duration:   0,
			Error:      err,
			StatusCode: 0,
		}
	}

	// Set headers
	for key, value := range requestConfig.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return Result{
			Duration:   duration,
			Error:      err,
			StatusCode: 0,
		}
	}

	// Discard and close the response body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	return Result{
		Duration:   duration,
		Error:      nil,
		StatusCode: resp.StatusCode,
	}
}
