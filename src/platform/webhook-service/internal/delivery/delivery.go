package delivery

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	maxAttempts    = 3
	initialBackoff = 2 * time.Second
)

// Dispatcher sends a webhook payload to the registered URL with retries.
type Dispatcher struct {
	client *http.Client
}

func NewDispatcher(timeout time.Duration) *Dispatcher {
	return &Dispatcher{client: &http.Client{Timeout: timeout}}
}

type Result struct {
	StatusCode int
	Attempt    int
	Success    bool
}

// Send posts payload to url, signing with secret if non-empty.
// Returns the last attempt result regardless of success.
func (d *Dispatcher) Send(ctx context.Context, url, secret string, payload []byte) Result {
	var last Result
	backoff := initialBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		code, err := d.post(ctx, url, secret, payload)
		last = Result{StatusCode: code, Attempt: attempt, Success: err == nil && code < 300}
		if last.Success {
			return last
		}
		if attempt < maxAttempts {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return last
			}
		}
	}
	return last
}

func (d *Dispatcher) post(ctx context.Context, url, secret string, payload []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-Webhook-Signature", sign(secret, payload))
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	if resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("non-2xx status %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
