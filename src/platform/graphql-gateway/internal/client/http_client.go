package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Doer is the interface used by resolver to call downstream services.
type Doer interface {
	Get(ctx context.Context, url string) ([]byte, int, error)
	Post(ctx context.Context, url string, body []byte) ([]byte, int, error)
}

// HTTPClient wraps the standard library http.Client and satisfies Doer.
type HTTPClient struct {
	client *http.Client
}

// New creates a new HTTPClient with the given request timeout.
func New(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs an HTTP GET to the given URL using the provided context.
// It returns the response body bytes, the HTTP status code, and any error.
func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("building GET request for %s: %w", url, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing GET request for %s: %w", url, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body from %s: %w", url, err)
	}
	return data, resp.StatusCode, nil
}

// Post performs an HTTP POST to the given URL with the provided JSON body using
// the provided context.  It returns the response body bytes, the HTTP status
// code, and any error.
func (c *HTTPClient) Post(ctx context.Context, url string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("building POST request for %s: %w", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing POST request for %s: %w", url, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body from %s: %w", url, err)
	}
	return data, resp.StatusCode, nil
}
