// Shared helpers for the ConvertIntoMP4 API examples. API reference: https://convertintomp4.com/api-docs
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	apiURL = strings.TrimRight(envOr("CIM4_API_URL", "https://api.convertintomp4.com"), "/")
	apiKey = os.Getenv("CIM4_API_KEY")
	// Sandbox mode (X-Sandbox: true): files up to 1 MB, 100 requests a month per key. CIM4_SANDBOX=0 for live mode.
	sandbox    = os.Getenv("CIM4_SANDBOX") != "0"
	sampleURL  = "https://convertintomp4.com/icon-192.png"
	sampleFile = filepath.Join("..", "fixtures", "sample.png")
	httpClient = &http.Client{Timeout: 330 * time.Second}
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type Task struct {
	ID        string          `json:"id"`
	Operation string          `json:"operation"`
	Status    string          `json:"status"`
	Progress  float64         `json:"progress"`
	Error     string          `json:"error,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
}

type Job struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Tasks  map[string]Task `json:"tasks"`
}

// ExportResult is the result of an export/url task.
type ExportResult struct {
	URL  string `json:"url"`
	File struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"file"`
}

// APIError: every /v1/jobs error has one shape: {"success": false, "error": {"code", "message", "details"?}}.
type APIError struct {
	Status     int
	Code       string
	Message    string
	RetryAfter string
}

func (e *APIError) Error() string { return fmt.Sprintf("%d %s: %s", e.Status, e.Code, e.Message) }

func setAuth(req *http.Request) {
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("User-Agent", "convertintomp4-api-examples/1.0 (go)")
	if sandbox {
		req.Header.Set("X-Sandbox", "true")
	}
}

// call sends a request. body is a JSON value, or a *Form for multipart uploads.
func call(method, path string, body any, out any) error {
	for attempt := 1; ; attempt++ {
		var reader io.Reader
		contentType := ""
		switch b := body.(type) {
		case nil:
		case *Form:
			buf, ct, err := b.encode()
			if err != nil {
				return err
			}
			reader, contentType = buf, ct
		default:
			raw, err := json.Marshal(b)
			if err != nil {
				return err
			}
			reader, contentType = bytes.NewReader(raw), "application/json"
		}
		req, err := http.NewRequest(method, apiURL+path, reader)
		if err != nil {
			return err
		}
		setAuth(req)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		data, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode == 429 && attempt < 4 {
			// Rate limited (sandbox: 1 POST a second). Wait and send the same request again.
			wait := time.Duration(1500*attempt) * time.Millisecond
			if s, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil {
				wait = time.Duration(s) * time.Second
			}
			time.Sleep(wait)
			continue
		}
		if res.StatusCode >= 400 {
			var e struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if json.Unmarshal(data, &e) != nil || e.Error.Code == "" {
				e.Error.Code, e.Error.Message = "HTTP_"+strconv.Itoa(res.StatusCode), strings.TrimSpace(string(data[:min(len(data), 200)]))
			}
			return &APIError{res.StatusCode, e.Error.Code, e.Error.Message, res.Header.Get("Retry-After")}
		}
		if out != nil && len(data) > 0 {
			return json.Unmarshal(data, out)
		}
		return nil
	}
}

// Form is a multipart body: text fields plus one file. It is encoded again on every attempt.
type Form struct {
	Fields    map[string]string
	FileField string
	FileName  string
	FileBytes []byte
}

func (f *Form) encode() (*bytes.Buffer, string, error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	for k, v := range f.Fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}
	part, err := w.CreateFormFile(f.FileField, f.FileName)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(f.FileBytes); err != nil {
		return nil, "", err
	}
	return buf, w.FormDataContentType(), w.Close()
}

// waitForJob waits for a job to settle. One /wait call holds at most 90 s; a longer job
// answers 408 TIMEOUT and we simply ask again.
func waitForJob(id string) (*Job, error) {
	deadline := time.Now().Add(10 * time.Minute)
	for {
		var job Job
		err := call("GET", "/v1/jobs/"+id+"/wait?timeout=90", nil, &job)
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 408 && time.Now().Before(deadline) {
			continue
		}
		return &job, err
	}
}

func exportResult(job *Job, task string) ExportResult {
	var r ExportResult
	_ = json.Unmarshal(job.Tasks[task].Result, &r)
	return r
}

func printTasks(job *Job) {
	for name, t := range job.Tasks {
		line := fmt.Sprintf("  %-12s %-16s %s", name, t.Operation, t.Status)
		if t.Error != "" {
			line += "  " + t.Error
		}
		fmt.Println(line)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
