// Package client is a minimal HTTP client for the MISP REST API.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client talks to a MISP instance over its REST API.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// Config holds the parameters needed to construct a Client.
type Config struct {
	URL       string
	APIKey    string
	Insecure  bool
	UserAgent string
	Timeout   time.Duration
}

// New returns a configured MISP client.
func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("misp: url is required")
	}
	if cfg.APIKey == "" {
		return nil, errors.New("misp: api_key is required")
	}

	u, err := url.Parse(strings.TrimRight(cfg.URL, "/"))
	if err != nil {
		return nil, fmt.Errorf("misp: invalid url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("misp: url must include scheme and host, got %q", cfg.URL)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //#nosec G402 -- opt-in
	}

	ua := cfg.UserAgent
	if ua == "" {
		ua = "terraform-provider-misp"
	}

	return &Client{
		baseURL:    u,
		apiKey:     cfg.APIKey,
		httpClient: &http.Client{Timeout: timeout, Transport: transport},
		userAgent:  ua,
	}, nil
}

// APIError is returned when MISP responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("misp: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// IsNotFound reports whether the error represents a 404 from MISP.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// do performs an HTTP request and decodes a JSON body into out (if non-nil).
// A nil body argument sends no request body.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("misp: marshal request: %w", err)
		}
		buf = bytes.NewReader(b)
	}

	u := *c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return fmt.Errorf("misp: build request: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("misp: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("misp: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       string(respBody),
		}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("misp: decode response: %w (body: %s)", err, truncate(string(respBody), 256))
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
