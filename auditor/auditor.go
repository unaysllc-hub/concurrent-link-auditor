package auditor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"statusCode"`
	DurationMS int64  `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}

type indexedResult struct {
	index  int
	result Result
}

func Check(ctx context.Context, client *http.Client, urls []string, workers int) []Result {
	if workers < 1 {
		workers = 1
	}
	if workers > len(urls) && len(urls) > 0 {
		workers = len(urls)
	}
	jobs := make(chan int)
	results := make(chan indexedResult)
	var group sync.WaitGroup

	worker := func() {
		defer group.Done()
		for index := range jobs {
			results <- indexedResult{index: index, result: checkOne(ctx, client, urls[index])}
		}
	}
	group.Add(workers)
	for range workers {
		go worker()
	}
	go func() {
		defer close(jobs)
		for index := range urls {
			select {
			case jobs <- index:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		group.Wait()
		close(results)
	}()

	ordered := make([]Result, len(urls))
	for item := range results {
		ordered[item.index] = item.result
	}
	return ordered
}

func checkOne(ctx context.Context, client *http.Client, rawURL string) Result {
	result := Result{URL: rawURL}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		result.Error = "invalid HTTP or HTTPS URL"
		return result
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	request.Header.Set("User-Agent", "Unays-Link-Auditor/1.0")
	started := time.Now()
	response, err := client.Do(request)
	result.DurationMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	result.StatusCode = response.StatusCode
	return result
}

func (result Result) Healthy() bool {
	return result.Error == "" && result.StatusCode >= 200 && result.StatusCode < 400
}

func (result Result) String() string {
	if result.Error != "" {
		return fmt.Sprintf("ERROR\t%s\t%s", result.URL, result.Error)
	}
	return fmt.Sprintf("%d\t%dms\t%s", result.StatusCode, result.DurationMS, result.URL)
}
