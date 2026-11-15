package main

import (
	"fmt"
	"strings"
	"time"
)

// 2. Check status with polling: GET /v1/jobs/{id} until the job is completed, failed or cancelled.
func checkStatusWithPolling() {
	var job Job
	must(call("POST", "/v1/jobs", map[string]any{"tasks": map[string]any{
		"import-logo": map[string]any{"operation": "import/url", "url": sampleURL},
		"to-webp":     map[string]any{"operation": "convert", "input": "import-logo", "output_format": "webp"},
		"export":      map[string]any{"operation": "export/url", "input": "to-webp"},
	}}, &job))
	fmt.Println("Created job", job.ID)

	deadline := time.Now().Add(10 * time.Minute)
	delay := time.Second
	for job.Status != "completed" && job.Status != "failed" && job.Status != "cancelled" {
		if time.Now().After(deadline) {
			must(fmt.Errorf("gave up after 10 minutes"))
		}
		time.Sleep(delay)
		delay = min(delay*3/2, 10*time.Second) // back off: most jobs finish in seconds, long videos take minutes
		must(call("GET", "/v1/jobs/"+job.ID, nil, &job))
		var p []string
		for _, t := range job.Tasks {
			p = append(p, fmt.Sprintf("%.0f%%", t.Progress))
		}
		fmt.Println(" ", job.Status, strings.Join(p, " "))
	}
	printTasks(&job)
	if job.Status == "completed" {
		fmt.Println("Download:", exportResult(&job, "export").URL)
	}
}
