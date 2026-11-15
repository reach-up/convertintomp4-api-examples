package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 5. Results and errors.
// Results: an export/url task returns {"url", "file": {"name", "size"}}; the URL is signed and expires.
// Errors: {"success": false, "error": {"code", "message", "details"?}} with the HTTP status.

// withRetry retries only what can succeed later: rate limits and server errors. Never a 4xx input error.
func withRetry(fn func() error) error {
	for i := 1; ; i++ {
		err := fn()
		var apiErr *APIError
		if err == nil || !errors.As(err, &apiErr) || !(apiErr.Status == 429 || apiErr.Status >= 500) || i == 4 {
			return err
		}
		wait := time.Duration(1<<i) * time.Second
		if s, e := strconv.Atoi(apiErr.RetryAfter); e == nil {
			wait = time.Duration(s) * time.Second
		}
		fmt.Println(" ", apiErr.Code, "retrying in", wait)
		time.Sleep(wait)
	}
}

func handlingResultsAndErrors() {
	// An input error: the task names an input that does not exist.
	err := call("POST", "/v1/jobs", map[string]any{"tasks": map[string]any{
		"to-jpg": map[string]any{"operation": "convert", "input": "nothing", "output_format": "jpg"},
	}}, nil)
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		fmt.Println("Expected error:", apiErr)
	}

	// A job that succeeds: download the result.
	var job Job
	must(withRetry(func() error {
		return call("POST", "/v1/jobs", map[string]any{"tasks": map[string]any{
			"import-logo": map[string]any{"operation": "import/url", "url": sampleURL},
			"to-png":      map[string]any{"operation": "thumbnail", "input": "import-logo", "output_format": "png", "options": map[string]any{"width": 64, "height": 64, "fit": "crop"}},
			"export":      map[string]any{"operation": "export/url", "input": "to-png"},
		}}, &job)
	}))
	var done *Job
	must(withRetry(func() (e error) { done, e = waitForJob(job.ID); return }))
	if done.Status != "completed" {
		// A failed task carries its own error; tasks that depended on it are cancelled.
		for name, t := range done.Tasks {
			if t.Error != "" {
				fmt.Printf("Task %s failed: %s\n", name, t.Error)
			}
		}
		os.Exit(1)
	}
	result := exportResult(done, "export")
	res, err := http.Get(result.URL)
	must(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		must(fmt.Errorf("download failed: %d", res.StatusCode))
	}
	data, err := io.ReadAll(res.Body)
	must(err)
	name := result.File.Name
	if name == "" {
		name = "result.png"
	}
	must(os.WriteFile(name, data, 0o644))
	fmt.Println("Saved", name, len(data), "bytes")
}
