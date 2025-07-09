package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// Result represents the outcome of an HTTP request
type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

// Report contains the final statistics of the load test
type Report struct {
	TotalRequests      int
	TotalDuration      time.Duration
	StatusCodes        map[int]int
	SuccessfulRequests int
	FailedRequests     int
	AverageTime        time.Duration
	MinTime            time.Duration
	MaxTime            time.Duration
}

func main() {
	// Parse command line arguments
	url := flag.String("url", "", "URL of the service to test")
	requests := flag.Int("requests", 100, "Total number of requests")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent requests")

	flag.Parse()

	// Validate URL parameter
	if *url == "" {
		fmt.Println("Error: URL is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate requests and concurrency parameters
	if *requests <= 0 {
		fmt.Println("Error: Number of requests must be greater than 0")
		os.Exit(1)
	}

	if *concurrency <= 0 || *concurrency > *requests {
		fmt.Println("Error: Concurrency must be greater than 0 and less than or equal to the number of requests")
		os.Exit(1)
	}

	fmt.Printf("Starting load test for %s\n", *url)
	fmt.Printf("Total requests: %d\n", *requests)
	fmt.Printf("Concurrency level: %d\n\n", *concurrency)

	// Run the load test
	report := runLoadTest(*url, *requests, *concurrency)

	// Print the report
	printReport(report)
}

func runLoadTest(url string, totalRequests, concurrency int) Report {
	// Create a channel to receive results
	resultChan := make(chan Result, totalRequests)

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup

	// Create a semaphore channel to limit concurrency
	semaphore := make(chan struct{}, concurrency)

	// Record start time
	startTime := time.Now()

	// Launch worker goroutines
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore

			// Make HTTP request and measure time
			start := time.Now()
			resp, err := http.Get(url)
			duration := time.Since(start)

			result := Result{
				Duration: duration,
				Error:    err,
			}

			if err == nil {
				result.StatusCode = resp.StatusCode
				resp.Body.Close()
			}

			resultChan <- result
		}()
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	report := Report{
		TotalRequests: totalRequests,
		StatusCodes:   make(map[int]int),
		MinTime:       time.Hour, // Initialize with a large value
	}

	var totalTime time.Duration

	for result := range resultChan {
		if result.Error != nil {
			report.FailedRequests++
			continue
		}

		report.StatusCodes[result.StatusCode]++
		totalTime += result.Duration

		if result.StatusCode == http.StatusOK {
			report.SuccessfulRequests++
		}

		// Update min and max times
		if result.Duration < report.MinTime {
			report.MinTime = result.Duration
		}
		if result.Duration > report.MaxTime {
			report.MaxTime = result.Duration
		}
	}

	// Calculate total duration and average time
	report.TotalDuration = time.Since(startTime)
	if totalRequests-report.FailedRequests > 0 {
		report.AverageTime = totalTime / time.Duration(totalRequests-report.FailedRequests)
	}

	return report
}

func printReport(report Report) {
	fmt.Println("=== Load Test Report ===")
	fmt.Printf("Total time: %v\n", report.TotalDuration)
	fmt.Printf("Total requests: %d\n", report.TotalRequests)
	fmt.Printf("Successful requests (HTTP 200): %d\n", report.SuccessfulRequests)
	fmt.Printf("Failed requests: %d\n", report.FailedRequests)
	fmt.Printf("Requests per second: %.2f\n", float64(report.TotalRequests)/report.TotalDuration.Seconds())
	fmt.Printf("Average response time: %v\n", report.AverageTime)
	fmt.Printf("Min response time: %v\n", report.MinTime)
	fmt.Printf("Max response time: %v\n", report.MaxTime)

	fmt.Println("\nStatus code distribution:")
	for code, count := range report.StatusCodes {
		fmt.Printf("  [%d]: %d responses\n", code, count)
	}
}
