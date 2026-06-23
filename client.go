package socialapis

import (
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

// DefaultBaseURL is the production socialapis.io REST endpoint.
const DefaultBaseURL = "https://api.socialapis.io"

// DefaultTimeout for HTTP requests when no per-call deadline is set
// via context.
const DefaultTimeout = 30 * time.Second

// HTTPDoer is the minimum interface the SDK needs from an HTTP client.
// stdlib's *http.Client satisfies this. Custom implementations (e.g.
// adding retry logic, metrics, or per-request tracing) can also be
// dropped in via WithHTTPClient.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Option configures a Facebook / Instagram / Account client. Use the
// `With...` functions below — they cover every configurable behaviour.
type Option func(*baseConfig)

type baseConfig struct {
	apiToken   string
	baseURL    string
	httpClient HTTPDoer
}

// WithBaseURL overrides the API base URL. Useful for staging or local
// mock servers.
func WithBaseURL(baseURL string) Option {
	return func(c *baseConfig) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient lets callers pass a custom HTTPDoer (typically an
// *http.Client with a different timeout or transport).
func WithHTTPClient(client HTTPDoer) Option {
	return func(c *baseConfig) {
		c.httpClient = client
	}
}

// newBaseConfig assembles a configured baseConfig from a token + options.
// Validates the token is non-empty.
func newBaseConfig(apiToken string, opts ...Option) (*baseConfig, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("socialapis: apiToken is required (get a free key at https://socialapis.io/auth/signup)")
	}
	c := &baseConfig{
		apiToken:   apiToken,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// get is the shared HTTP driver for every endpoint method. It:
//   - builds the full URL from base + path + params
//   - sets the auth header
//   - maps HTTP status codes to typed errors
//   - parses the JSON body into the caller's *any / *map / *struct
func (c *baseConfig) get(ctx context.Context, path string, params url.Values, out any) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("socialapis: path must start with '/', got %q", path)
	}

	fullURL := c.baseURL + path
	if encoded := params.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return &ConnectionError{Message: "failed to build request", Cause: err}
	}
	req.Header.Set("x-api-token", c.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &ConnectionError{Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ConnectionError{Message: "failed to read response body", Cause: err}
	}

	if resp.StatusCode >= 400 {
		return c.translateError(resp, body)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return &ConnectionError{Message: "failed to decode JSON response", Cause: err}
	}
	return nil
}

// translateError maps an HTTP error response to one of the typed
// SDK errors (AuthenticationError, RateLimitError, etc.).
func (c *baseConfig) translateError(resp *http.Response, body []byte) error {
	parsed := map[string]any{}
	_ = json.Unmarshal(body, &parsed) // best-effort; empty map is fine if non-JSON

	message := extractMessage(parsed)
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	if message == "" {
		message = resp.Status
	}

	base := APIError{
		Message:    message,
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("x-request-id"),
		Body:       parsed,
	}

	switch resp.StatusCode {
	case 401:
		return &AuthenticationError{APIError: base}
	case 402:
		return &InsufficientCreditsError{APIError: base}
	case 429:
		retryAfter := 0.0
		if v := resp.Header.Get("retry-after"); v != "" {
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				retryAfter = parsed
			}
		}
		return &RateLimitError{APIError: base, RetryAfterSeconds: retryAfter}
	}
	if resp.StatusCode >= 500 {
		return &APIServerError{APIError: base}
	}
	return &BadRequestError{APIError: base}
}

// extractMessage pulls a human-readable error from the API's response
// envelope. The API uses different keys across endpoints — try the
// most common ones in order.
func extractMessage(body map[string]any) string {
	for _, key := range []string{"error", "message", "detail"} {
		v, ok := body[key]
		if !ok {
			continue
		}
		if s, ok := v.(string); ok && s != "" {
			return s
		}
		if nested, ok := v.(map[string]any); ok {
			if s, ok := nested["message"].(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// ---------------------------------------------------------------------
// Identifier normalisation helpers
// ---------------------------------------------------------------------

// asFacebookURL normalises a slug or full URL to a canonical Facebook URL.
//
//	"EngenSA"                          → "https://www.facebook.com/EngenSA"
//	"https://www.facebook.com/EngenSA" → unchanged
func asFacebookURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("socialapis: identifier is required")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, nil
	}
	return "https://www.facebook.com/" + strings.TrimLeft(trimmed, "/"), nil
}

// asFacebookGroupURL normalises a Group identifier (slug, numeric id,
// or full URL) to a canonical Facebook group URL.
func asFacebookGroupURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("socialapis: group identifier is required")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, nil
	}
	return "https://www.facebook.com/groups/" + strings.TrimLeft(trimmed, "/"), nil
}

// asInstagramURL normalises an Instagram identifier to a canonical
// profile URL.
func asInstagramURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("socialapis: identifier is required")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, nil
	}
	return "https://www.instagram.com/" + strings.Trim(trimmed, "/"), nil
}

// mergeParams takes a primary {key: value} map plus an `extra` map of
// arbitrary forward-compat kwargs and returns url.Values. Nil / zero
// values are dropped. Used by every endpoint method.
func mergeParams(primary map[string]string, extra map[string]any) url.Values {
	values := url.Values{}
	for k, v := range primary {
		if v != "" {
			values.Set(k, v)
		}
	}
	for k, v := range extra {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			if val != "" {
				values.Set(k, val)
			}
		default:
			values.Set(k, fmt.Sprint(v))
		}
	}
	return values
}
