// Official dogabot REST API SDK for Go.
//
// Module: github.com/dogabot/dogabot-sdk-go
package dogabot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.dogabot.com"
const sdkVersion = "0.1.0"

// Client is the dogabot REST API client.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	UserAgent  string
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the API base URL.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) { c.BaseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) { c.HTTPClient = h }
}

// WithMaxRetries sets retries for 429/5xx.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) { c.MaxRetries = n }
}

// New creates a Client. apiKey may be empty if DOGABOT_API_KEY is set via env in callers;
// pass the key explicitly for clarity.
func New(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
		MaxRetries: 2,
		UserAgent:  "dogabot-sdk-go/" + sdkVersion,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type requestOptions struct {
	query           url.Values
	headers         map[string]string
	body            any
	idempotencyKey  string
	write           bool
}

// RequestOption mutates a single request.
type RequestOption func(*requestOptions)

// WithQuery sets query parameters.
func WithQuery(q url.Values) RequestOption {
	return func(r *requestOptions) { r.query = q }
}

// WithIdempotencyKey sets Idempotency-Key (required for writes).
func WithIdempotencyKey(key string) RequestOption {
	return func(r *requestOptions) { r.idempotencyKey = key }
}

// WithHeaders merges extra headers.
func WithHeaders(h map[string]string) RequestOption {
	return func(r *requestOptions) { r.headers = h }
}

// APIError is a non-2xx API response.
type APIError struct {
	StatusCode int
	Message    string
	Body       map[string]any
	Header     http.Header
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func (c *Client) do(ctx context.Context, method, path string, ro *requestOptions) (any, error) {
	if ro == nil {
		ro = &requestOptions{}
	}
	if c.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if ro.write && ro.idempotencyKey == "" {
		return nil, fmt.Errorf("Idempotency-Key is required for write operation %s %s", method, path)
	}

	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base + path)
	if err != nil {
		return nil, err
	}
	if ro.query != nil {
		u.RawQuery = ro.query.Encode()
	}

	var bodyReader io.Reader
	if ro.body != nil {
		b, err := json.Marshal(ro.body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	attempt := 0
	for {
		req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Accept", "application/json")
		ua := c.UserAgent
		if ua == "" {
			ua = "dogabot-sdk-go/" + sdkVersion
		}
		req.Header.Set("User-Agent", ua)
		if ro.body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if ro.idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", ro.idempotencyKey)
		}
		for k, v := range ro.headers {
			req.Header.Set(k, v)
		}

		hc := c.HTTPClient
		if hc == nil {
			hc = http.DefaultClient
		}
		res, err := hc.Do(req)
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		if res.StatusCode >= 200 && res.StatusCode < 300 {
			if len(raw) == 0 {
				return nil, nil
			}
			var out any
			if err := json.Unmarshal(raw, &out); err != nil {
				return string(raw), nil
			}
			return out, nil
		}

		retryable := res.StatusCode == 429 || res.StatusCode >= 500
		maxRetries := c.MaxRetries
		if retryable && attempt < maxRetries {
			attempt++
			wait := time.Duration(1<<attempt) * time.Second
			if ra := res.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			if ro.body != nil {
				b, _ := json.Marshal(ro.body)
				bodyReader = bytes.NewReader(b)
			}
			continue
		}

		var bodyMap map[string]any
		_ = json.Unmarshal(raw, &bodyMap)
		msg := fmt.Sprintf("HTTP %d %s %s", res.StatusCode, method, path)
		if bodyMap != nil {
			if e, ok := bodyMap["error"].(string); ok && e != "" {
				msg = e
			}
		}
		return nil, &APIError{StatusCode: res.StatusCode, Message: msg, Body: bodyMap, Header: res.Header}
	}
}
