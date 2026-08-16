# ⚡ CLI Benchmarking Tool (`go-bench`)

A fast, lightweight, and concurrent HTTP benchmarking command-line tool written in Go. Designed to stress-test and evaluate the performance of HTTP web servers, APIs, and microservices with granular metrics.

---

## ✨ Features

- **🚀 High Concurrency**: Efficient worker pool architecture running across Goroutines.
- **🔄 Any HTTP Method**: Full support for `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`, and more.
- **📋 Custom Headers**: Repeatable `-H` / `-header` flag to pass authorization tokens, content types, or custom metadata.
- **📝 Request Body / Payloads**: Pass raw JSON, XML, or plain text payloads with `-d` / `-data` / `-body`.
- **⏱️ Per-Request Timeouts**: Configurable timeout (`-t` / `-timeout`) to prevent hanging connections.
- **🔌 Optimized Connection Pooling**: Tuned `http.Transport` connection reuse preventing socket exhaustion at high RPS.
- **📊 Rich Performance Metrics**:
  - Total elapsed time & throughput (Requests Per Second - RPS).
  - Success vs. Failed requests count.
  - Latency statistics (Average, Minimum, Maximum).
  - Detailed **HTTP Status Code distribution** (e.g. `200 OK`, `400 Bad Request`, `404 Not Found`).
  - Network and connection **Error breakdown**.

---

## 📦 Installation & Setup

### Prerequisites
- [Go](https://go.dev/dl/) `1.20+` installed.

### Build from Source
```bash
# Clone the repository
git clone https://github.com/mbrik/CLI-Benchmarking-Tool.git
cd CLI-Benchmarking-Tool/go-bench

# Build executable
go build -o go-bench .
```

On Windows, this produces `go-bench.exe`.

---

## 🛠️ CLI Flags & Usage

```text
Usage: go-bench [options]

Flags:
  -url string       Target URL to benchmark (default "http://localhost:8080")
  -m, -method       HTTP method: GET, POST, PUT, DELETE, PATCH, etc. (default "GET")
  -n int            Total number of requests to execute (default 1000)
  -c int            Number of concurrent workers (default 10)
  -d, -data, -body  Request payload / data string (default "")
  -H, -header       Custom HTTP header (repeatable for multiple headers)
  -t, -timeout      Per-request timeout duration, e.g. 5s, 10s, 1m (default 10s)
  -help             Show help documentation
```

---

## 🚀 Examples

### 1. Basic GET Benchmark
Benchmark a GET endpoint with 2,000 total requests across 20 concurrent workers:
```bash
./go-bench -url "http://localhost:8080/get?key=username" -n 2000 -c 20
```

### 2. POST Benchmark with JSON Body & Headers
Send POST requests with a JSON body and custom authentication headers:
```bash
./go-bench -url "http://localhost:8080/api/users" \
  -m POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my_secret_token" \
  -d '{"name": "Alice", "role": "admin"}' \
  -n 1000 \
  -c 25
```

### 3. PUT Request with Custom Timeout
```bash
./go-bench -url "http://localhost:8080/api/resource/1" \
  -m PUT \
  -d '{"status": "active"}' \
  -t 5s \
  -n 500 \
  -c 10
```

### 4. DELETE Benchmark
```bash
./go-bench -url "http://localhost:8080/api/items/42" -m DELETE -n 200 -c 10
```

---

## 📊 Sample Output

```text
🎯 Benchmarking Target: [POST] http://localhost:8080/set?key=username&value=test
📦 Total Requests: 50000 | Concurrency: 20 | Timeout: 10s
⏳ Running benchmark, please wait...

==================================
📊 BENCHMARK RESULTS SUMMARY
==================================
Total Time Taken   : 39.5860176s
Successful Requests: 50000
Failed Requests    : 0
Requests Per Sec   : 1263.07 req/sec
Average Latency    : 15.801105ms
Minimum Latency    : 968.6µs
Maximum Latency    : 35.9232ms
----------------------------------
📈 Status Codes Distribution:
   [200 OK]: 50000 responses
==================================
```

When network errors or invalid endpoints occur:
```text
==================================
📊 BENCHMARK RESULTS SUMMARY
==================================
Total Time Taken   : 38.2996ms
Successful Requests: 0
Failed Requests    : 500
Requests Per Sec   : 13054.97 req/sec
Average Latency    : 1.309235ms
Minimum Latency    : 0s
Maximum Latency    : 12.7131ms
----------------------------------
📈 Status Codes Distribution:
   [400 Bad Request]: 500 responses
----------------------------------
❌ Error Breakdown:
   - dial tcp 127.0.0.1:8080: connectex: No connection could be made... : 10 occurrences
==================================
```

---

## 🧪 Running Unit Tests

Run the test suite:
```bash
go test -v ./...
```

---

## 📁 Project Structure

```text
go-bench/
├── cmd/
│   └── root.go          # CLI flag parsing, pre-execution banner, and report rendering
├── internal/
│   ├── models.go        # (or within worker.go) Request & Benchmark configuration types
│   ├── runner.go        # Concurrent worker pool and job distribution
│   ├── runner_test.go   # Automated tests with httptest server
│   ├── stats.go         # Latency, status codes, and error metrics calculation
│   └── worker.go        # HTTP client setup, connection pooling, and request execution
├── main.go              # Entry point invoking cmd.Execute()
├── go.mod               # Module configuration
└── README.md            # Documentation
```

---

## 📄 License
This project is open source and available under the [MIT License](LICENSE).
