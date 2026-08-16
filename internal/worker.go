package internal
import(
	"net/http"
	"time"
)
// Result holds the metrics for a single HTTP request execution
type Result struct{
	Duration time.Duration
	Error error
	StatusCode int 
}
// SendRequest executes a single HTTP GET request and measures its duration 
func SendRequest(url string) Result{
	start := time.Now()

	// send the HTTp GET request to the target URL
	resp, err:=http.Get(url)
	duration := time.Since(start)

	if err !=nil{
		return Result{
			Duration: duration,
			Error: err,
			StatusCode:0,
		}
	}

	defer resp.Body.Close()

	// Retuen the performance and status metrics
	return Result{
		Duration: duration,
		Error: nil,
		StatusCode: resp.StatusCode,
	}
	
}

