package internal

import (
	"sync"
)

// RunBenchmark executes totalRequests concurrently using specified number of workers
func RunBenchmark(url string, totalRequests, concurrency int)[]Result{
	// Creare a buffered channel to collect results safely
	results := make(chan Result, totalRequests)
	var wg sync.WaitGroup

	// Calculate how many requests each worker should handle
	requestsPerWorker := totalRequests / concurrency

	// Spawn worker goroutines
	for i:=0;i <concurrency;i++{
		wg.Add(1)
		go func(){
			defer wg.Done()
			for j:=0;j < requestsPerWorker;j++{
				res:= SendRequest(url)
				results <-res
			}	
	}()

	}
	wg.Wait()
	close(results)
	
	// Collect all results from the channel into a slice
	var allResults []Result
	for res := range results {
		allResults = append(allResults, res)
	}
	
	return allResults

}
	
