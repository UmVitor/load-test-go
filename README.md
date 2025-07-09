# Load Balance - Web Service Load Testing Tool

A command-line tool written in Go for load testing web services. This tool allows you to specify the number of requests and concurrency level to test the performance of any web service.

## Features

- Configurable number of total requests
- Adjustable concurrency level
- Detailed performance report including:
  - Total execution time
  - Request success/failure counts
  - HTTP status code distribution
  - Response time statistics (min, max, average)

## Usage

### Command Line Parameters

- `--url`: URL of the service to test (required)
- `--requests`: Total number of requests to make (default: 100)
- `--concurrency`: Number of concurrent requests (default: 10)

### Examples

Run directly with Go:

```bash
go run main.go --url=https://example.com --requests=1000 --concurrency=10
```

Or build and run the binary:

```bash
go build -o load-balancer
./load-balancer --url=https://example.com --requests=1000 --concurrency=10
```

### Docker Usage

Build the Docker image:

```bash
docker build -t load-balancer .
```

Run the load test using Docker:

```bash
docker run load-balancer --url=https://example.com --requests=1000 --concurrency=10
```

## Sample Output

```
Starting load test for https://example.com
Total requests: 1000
Concurrency level: 10

=== Load Test Report ===
Total time: 5.721s
Total requests: 1000
Successful requests (HTTP 200): 998
Failed requests: 2
Requests per second: 174.79
Average response time: 56.9ms
Min response time: 42.1ms
Max response time: 312.5ms

Status code distribution:
  [200]: 998 responses
  [500]: 2 responses
```

## Building from Source

```bash
git clone <repository-url>
cd load-balance
go build
```

## Requirements

- Go 1.18 or higher
