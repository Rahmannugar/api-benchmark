package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

func TestBenchmarkConfigValidate(t *testing.T) {
	validConfig := func() config.BenchmarkConfig {
		return config.BenchmarkConfig{
			Request: config.RequestConfig{
				Method:  "GET",
				URL:     "https://example.com/api/items?limit=10",
				Timeout: 10 * time.Second,
			},
			TotalRequests: 100,
			Concurrency:   10,
		}
	}

	tests := []struct {
		name        string
		configure   func(*config.BenchmarkConfig)
		wantErrText string
	}{
		{name: "valid configuration"},
		{
			name: "total requests must be positive",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.TotalRequests = 0
			},
			wantErrText: "total requests",
		},
		{
			name: "concurrency must be positive",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Concurrency = 0
			},
			wantErrText: "concurrency must be greater than zero",
		},
		{
			name: "concurrency cannot exceed requests",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Concurrency = cfg.TotalRequests + 1
			},
			wantErrText: "concurrency cannot exceed total requests",
		},
		{
			name: "method is required",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Request.Method = " "
			},
			wantErrText: "HTTP method cannot be empty",
		},
		{
			name: "method must be valid",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Request.Method = "INVALID METHOD"
			},
			wantErrText: "invalid HTTP method",
		},
		{
			name: "URL scheme is required",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Request.URL = "example.com/api"
			},
			wantErrText: "URL scheme must be http or https",
		},
		{
			name: "URL scheme must be HTTP",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Request.URL = "ftp://example.com/file"
			},
			wantErrText: "URL scheme must be http or https",
		},
		{
			name: "URL host is required",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Request.URL = "http:///api/items"
			},
			wantErrText: "URL must include a host",
		},
		{
			name: "timeout must be positive",
			configure: func(cfg *config.BenchmarkConfig) {
				cfg.Request.Timeout = 0
			},
			wantErrText: "timeout must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			if tt.configure != nil {
				tt.configure(&cfg)
			}

			err := cfg.Validate()
			if tt.wantErrText == "" {
				if err != nil {
					t.Fatalf("expected valid configuration, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErrText)
			}
			if !strings.Contains(err.Error(), tt.wantErrText) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErrText, err)
			}
		})
	}
}
