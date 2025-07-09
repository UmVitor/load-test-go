package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)


type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

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
	url := flag.String("url", "", "URL of the service to test")
	requests := flag.Int("requests", 100, "Total number of requests")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent requests")

	flag.Parse()

	if *url == "" {
		fmt.Println("Error: URL is required")
		flag.Usage()
		os.Exit(1)
	}

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
	
	report := runLoadTest(*url, *requests, *concurrency)

	printReport(report)
}

func runLoadTest(url string, totalRequests, concurrency int) Report {
	resultChan := make(chan Result, totalRequests)
	
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, concurrency)

	startTime := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
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

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	report := Report{
		TotalRequests: totalRequests,
		StatusCodes:   make(map[int]int),
		MinTime:       time.Hour,
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

		if result.Duration < report.MinTime {
			report.MinTime = result.Duration
		}
		if result.Duration > report.MaxTime {
			report.MaxTime = result.Duration
		}
	}

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
