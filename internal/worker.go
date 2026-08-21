package internal

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

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
func SendRequest(client *http.Client, requestConfig *config.RequestConfig) RequestResult {
	var bodyReader io.Reader
	if len(requestConfig.Body) > 0 {
		bodyReader = bytes.NewReader(requestConfig.Body)
	}

	req, err := http.NewRequest(requestConfig.Method, requestConfig.URL, bodyReader)
	if err != nil {
		return RequestResult{
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
	if err != nil {
		return RequestResult{
			Duration:   time.Since(start),
			Error:      err,
			StatusCode: 0,
		}
	}

	_, readErr := io.Copy(io.Discard, resp.Body)
	closeErr := resp.Body.Close()
	duration := time.Since(start)
	if err := errors.Join(readErr, closeErr); err != nil {
		return RequestResult{
			Duration:   duration,
			Error:      err,
			StatusCode: resp.StatusCode,
		}
	}

	return RequestResult{
		Duration:   duration,
		Error:      nil,
		StatusCode: resp.StatusCode,
	}
}
