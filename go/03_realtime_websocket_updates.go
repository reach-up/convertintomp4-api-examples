package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

// 3. Realtime updates over a WebSocket instead of polling.
// wss://api.convertintomp4.com/ws/user/jobs streams every job of the key's account:
// {"type": "started" | "progress" | "completed" | "failed", "jobId", "progress"?, "outputUrl"?, "error"?}.
// For multi-task jobs, also pass "webhook_url" when you create the job: you get one signed
// job.finished / job.failed POST when the whole job settles (see verifyWebhook below).
func realtimeWebsocketUpdates() {
	header := http.Header{}
	header.Set("X-Api-Key", apiKey)
	header.Set("User-Agent", "convertintomp4-api-examples/1.0 (go)")
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(apiURL, "http", "ws", 1)+"/ws/user/jobs", header)
	must(err)
	defer conn.Close()
	var hello map[string]any
	must(conn.ReadJSON(&hello)) // {"type": "connected"}

	// Start the conversion only once the socket is open, so no event is missed.
	file, err := os.ReadFile(sampleFile)
	must(err)
	var created struct {
		JobID string `json:"jobId"`
	}
	must(call("POST", "/v1/convert", &Form{Fields: map[string]string{"targetFormat": "jpg"}, FileField: "file", FileName: "sample.png", FileBytes: file}, &created))
	fmt.Println("Created conversion", created.JobID)

	for {
		var msg struct {
			Type      string   `json:"type"`
			JobID     string   `json:"jobId"`
			Progress  *float64 `json:"progress"`
			OutputURL string   `json:"outputUrl"`
			Error     string   `json:"error"`
		}
		must(conn.ReadJSON(&msg))
		if msg.JobID != created.JobID {
			continue // other jobs of the same account
		}
		if msg.Progress != nil {
			fmt.Printf("  %s %.0f%%\n", msg.Type, *msg.Progress)
		} else {
			fmt.Println(" ", msg.Type)
		}
		switch msg.Type {
		case "completed":
			fmt.Println("Download:", apiURL+msg.OutputURL)
			return
		case "failed":
			fmt.Println("Failed:", msg.Error)
			return
		}
	}
}

// verifyWebhook checks X-ConvertIntoMP4-Signature on the raw body before you trust a webhook.
// The secret comes from GET /v1/jobs/webhook-secret.
func verifyWebhook(rawBody []byte, signatureHeader, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signatureHeader))
}

var _ = verifyWebhook
