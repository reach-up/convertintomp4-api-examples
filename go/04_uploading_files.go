package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

// 4. Uploading files.
// A) One call: POST /v1/convert with the file as multipart form data.
// B) In a job: an import/upload task returns a form URL; send the file there and the job continues.
func uploadingFiles() {
	file, err := os.ReadFile(sampleFile)
	must(err)

	// A) one call
	var conversion struct {
		JobID     string `json:"jobId"`
		Status    string `json:"status"`
		StatusURL string `json:"statusUrl"`
	}
	must(call("POST", "/v1/convert", &Form{Fields: map[string]string{"targetFormat": "webp"}, FileField: "file", FileName: "sample.png", FileBytes: file}, &conversion))
	fmt.Println("A) conversion", conversion.JobID, conversion.Status, "status at", conversion.StatusURL)

	// B) upload inside a job
	var job Job
	must(call("POST", "/v1/jobs", map[string]any{"tasks": map[string]any{
		"upload": map[string]any{"operation": "import/upload", "filename": "sample.png"},
		"to-jpg": map[string]any{"operation": "convert", "input": "upload", "output_format": "jpg"},
		"export": map[string]any{"operation": "export/url", "input": "to-jpg"},
	}}, &job))
	var upload struct {
		Form struct {
			URL    string `json:"url"`
			Method string `json:"method"`
			Field  string `json:"field"`
		} `json:"form"`
	}
	must(json.Unmarshal(job.Tasks["upload"].Result, &upload))
	u, err := url.Parse(upload.Form.URL)
	must(err)
	must(call(upload.Form.Method, u.Path, &Form{FileField: upload.Form.Field, FileName: "sample.png", FileBytes: file}, nil))
	fmt.Println("B) uploaded to job", job.ID)

	done, err := waitForJob(job.ID)
	must(err)
	printTasks(done)
	fmt.Println("Download:", exportResult(done, "export").URL)
}
