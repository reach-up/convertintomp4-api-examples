package main

import "fmt"

// 6. A complex job in one request:
//
//	page 1: a capture of a web page as PDF
//	page 2: an image converted to PDF
//	merge both into one PDF, make a 200 px thumbnail of the image,
//	pack the PDF and the thumbnail into a zip and export it.
func complexJob() {
	var job Job
	must(call("POST", "/v1/jobs", map[string]any{"tag": "example-06", "tasks": map[string]any{
		"page":     map[string]any{"operation": "capture-website", "url": "https://convertintomp4.com/api-docs", "output_format": "pdf", "options": map[string]any{"page_format": "A4", "print_background": true}},
		"logo":     map[string]any{"operation": "import/url", "url": sampleURL},
		"logo-pdf": map[string]any{"operation": "convert", "input": "logo", "output_format": "pdf"},
		"merged":   map[string]any{"operation": "merge", "input": []string{"page", "logo-pdf"}, "output_format": "pdf"},
		"thumb":    map[string]any{"operation": "thumbnail", "input": "logo", "output_format": "jpg", "options": map[string]any{"width": 200, "height": 200, "fit": "max"}},
		"package":  map[string]any{"operation": "archive/create", "input": []string{"merged", "thumb"}, "output_format": "zip"},
		"export":   map[string]any{"operation": "export/url", "input": "package"},
	}}, &job))
	fmt.Println("Created job", job.ID, "with", len(job.Tasks), "tasks")

	done, err := waitForJob(job.ID)
	must(err)
	fmt.Println("Job", done.Status)
	printTasks(done)
	if done.Status == "completed" {
		fmt.Println("Zip:", exportResult(done, "export").URL)
	}
}
