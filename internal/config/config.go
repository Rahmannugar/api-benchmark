package config

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestConfig contains the settings for one HTTP request.
type RequestConfig struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

// BenchmarkConfig contains the settings for a complete benchmark run.
type BenchmarkConfig struct {
	Request       RequestConfig
	TotalRequests int
	Concurrency   int
}

// Validate checks configuration that must be valid before workers start.
func (c BenchmarkConfig) Validate() error {
	if c.TotalRequests <= 0 {
		return fmt.Errorf("total requests must be greater than zero")
	}
	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than zero")
	}
	if c.Concurrency > c.TotalRequests {
		return fmt.Errorf("concurrency cannot exceed total requests")
	}
	if err := c.Request.Validate(); err != nil {
		return fmt.Errorf("request configuration: %w", err)
	}

	return nil
}

// Validate checks request settings before any network work begins.
func (c RequestConfig) Validate() error {
	method := strings.TrimSpace(c.Method)
	if method == "" {
		return fmt.Errorf("HTTP method cannot be empty")
	}

	parsedURL, err := url.Parse(strings.TrimSpace(c.URL))
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}
	if _, err := http.NewRequest(method, parsedURL.String(), nil); err != nil {
		return fmt.Errorf("invalid HTTP method or URL: %w", err)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	return nil
}
