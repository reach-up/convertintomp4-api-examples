package main

import "fmt"

// 1. Tasks and jobs: one request describes the whole pipeline.
// A job is a set of named tasks. A task reads the output of another task through "input".
func tasksAndJobs() {
	var job Job
	must(call("POST", "/v1/jobs", map[string]any{
		"tag": "example-01",
		"tasks": map[string]any{
			"import-logo": map[string]any{"operation": "import/url", "url": sampleURL},
			"to-jpg":      map[string]any{"operation": "convert", "input": "import-logo", "output_format": "jpg", "options": map[string]any{"background": "#ffffff"}},
			"export":      map[string]any{"operation": "export/url", "input": "to-jpg"},
		},
	}, &job))
	fmt.Println("Created job", job.ID, "status", job.Status)
	printTasks(&job)

	// Block until the job settles (90 s per call, repeated). Example 02 polls instead, example 03 listens.
	done, err := waitForJob(job.ID)
	must(err)
	fmt.Println("Job", done.ID, done.Status)
	printTasks(done)
	fmt.Println("Download:", exportResult(done, "export").URL)
}
