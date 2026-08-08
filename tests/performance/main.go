package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/api/v1/catalog/products?pageSize=24", "GET target")
	duration := flag.Duration("duration", 15*time.Second, "test duration")
	concurrency := flag.Int("concurrency", 20, "parallel clients")
	maxP95 := flag.Duration("max-p95", 250*time.Millisecond, "maximum accepted p95")
	maxErrorRate := flag.Float64("max-error-rate", 0.01, "maximum accepted error ratio")
	flag.Parse()
	if *duration <= 0 || *concurrency < 1 || *concurrency > 500 || *maxP95 <= 0 || *maxErrorRate < 0 || *maxErrorRate > 1 {
		fmt.Fprintln(os.Stderr, "invalid performance arguments")
		os.Exit(2)
	}
	transport := &http.Transport{MaxIdleConns: *concurrency * 2, MaxIdleConnsPerHost: *concurrency, MaxConnsPerHost: *concurrency}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	defer transport.CloseIdleConnections()
	for range 10 {
		if err := request(context.Background(), client, *url); err != nil {
			fmt.Fprintf(os.Stderr, "warmup failed: %v\n", err)
			os.Exit(1)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()
	var failures atomic.Int64
	latencies := make([]time.Duration, 0, 10000)
	var mu sync.Mutex
	var workers sync.WaitGroup
	started := time.Now()
	for range *concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for ctx.Err() == nil {
				requestStarted := time.Now()
				err := request(ctx, client, *url)
				elapsed := time.Since(requestStarted)
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					failures.Add(1)
				}
				mu.Lock()
				latencies = append(latencies, elapsed)
				mu.Unlock()
			}
		}()
	}
	workers.Wait()
	elapsed := time.Since(started)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	if len(latencies) == 0 {
		fmt.Fprintln(os.Stderr, "no completed requests")
		os.Exit(1)
	}
	failed := failures.Load()
	errorRate := float64(failed) / float64(len(latencies))
	p50, p95, p99 := percentile(latencies, 0.50), percentile(latencies, 0.95), percentile(latencies, 0.99)
	fmt.Printf("requests=%d failures=%d error_rate=%.4f rps=%.1f p50=%s p95=%s p99=%s\n",
		len(latencies), failed, errorRate, float64(len(latencies))/elapsed.Seconds(), p50, p95, p99)
	if errorRate > *maxErrorRate || p95 > *maxP95 {
		fmt.Fprintf(os.Stderr, "baseline failed: max_error_rate=%.4f max_p95=%s\n", *maxErrorRate, maxP95.String())
		os.Exit(1)
	}
}

func request(ctx context.Context, client *http.Client, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 2<<20))
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", response.StatusCode)
	}
	return nil
}

func percentile(values []time.Duration, quantile float64) time.Duration {
	index := int(float64(len(values)-1) * quantile)
	return values[index]
}
