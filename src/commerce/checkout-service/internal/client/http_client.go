package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is a thin wrapper around net/http.Client with context support.
type HTTPClient struct {
	client *http.Client
}

// New creates a new HTTPClient with the provided request timeout.
func New(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs an HTTP GET against url and returns the response body bytes.
func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http_client: build GET request: %w", err)
	}
	return c.do(req)
}

// Post performs an HTTP POST against url with the given JSON body and returns
// the response body bytes.
func (c *HTTPClient) Post(ctx context.Context, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("http_client: build POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// do executes the request and returns the response body or an error.
func (c *HTTPClient) do(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http_client: execute request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http_client: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http_client: upstream returned %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}
