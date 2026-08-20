package internal_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mbrik/CLI-Benchmarking-Tool/internal"
	"github.com/mbrik/CLI-Benchmarking-Tool/internal/config"
)

func TestSendRequestCustomMethodAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("X-Custom-Header") != "TestValue" {
			t.Errorf("expected X-Custom-Header to be TestValue, got %s", r.Header.Get("X-Custom-Header"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != `{"hello":"world"}` {
			t.Errorf("unexpected body: %s, err: %v", string(body), err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := internal.NewHTTPClient(5, 5*time.Second)
	reqCfg := config.RequestConfig{
		Method: "POST",
		URL:    server.URL,
		Headers: map[string]string{
			"X-Custom-Header": "TestValue",
		},
		Body:    []byte(`{"hello":"world"}`),
		Timeout: 5 * time.Second,
	}

	result := internal.SendRequest(client, &reqCfg)
	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
}

func TestRunBenchmark(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	benchCfg := config.BenchmarkConfig{
		Request: config.RequestConfig{
			Method:  "GET",
			URL:     server.URL,
			Timeout: 5 * time.Second,
		},
		TotalRequests: 105,
		Concurrency:   10,
	}

	results := internal.RunBenchmark(benchCfg)
	if len(results) != 105 {
		t.Fatalf("expected 105 results, got %d", len(results))
	}
	if atomic.LoadInt64(&requestCount) != 105 {
		t.Fatalf("expected server to receive 105 requests, got %d", requestCount)
	}

	summary := internal.CalculateStats(results, 100*time.Millisecond)
	if summary.SuccessfulRequests != 105 {
		t.Fatalf("expected 105 successful requests, got %d", summary.SuccessfulRequests)
	}
	if summary.FailedRequests != 0 {
		t.Fatalf("expected 0 failed requests, got %d", summary.FailedRequests)
	}
	if summary.StatusCodes[http.StatusOK] != 105 {
		t.Fatalf("expected 105 HTTP 200 responses, got %d", summary.StatusCodes[http.StatusOK])
	}
}
