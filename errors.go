package socialapis

import "fmt"

// Error is the base interface that every SDK-originating error satisfies.
// Callers can catch broadly with `errors.As(err, &socialapis.Error{...})`
// or narrowly with the typed subtypes below.
type Error interface {
	error
	socialapisError() // sealed marker
}

// APIError is returned for any HTTP error response from the API.
// Specific 4xx/5xx flavours are typed below — use errors.As to dispatch.
type APIError struct {
	Message    string
	StatusCode int
	RequestID  string
	Body       map[string]any
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("socialapis: %d %s (request_id=%s)", e.StatusCode, e.Message, e.RequestID)
	}
	return fmt.Sprintf("socialapis: %d %s", e.StatusCode, e.Message)
}

func (e *APIError) socialapisError() {}

// AuthenticationError is 401 — the API token is invalid or missing.
// Not retryable; the caller has to fix the token.
type AuthenticationError struct{ APIError }

func (e *AuthenticationError) Error() string {
	return "socialapis: authentication failed: " + e.APIError.Error()
}

// InsufficientCreditsError is 402 — credit balance exhausted.
// Retryable after refill / upgrade. Tracked as a distinct error so paid
// integrations can auto-top-up on this signal.
type InsufficientCreditsError struct{ APIError }

func (e *InsufficientCreditsError) Error() string {
	return "socialapis: insufficient credits: " + e.APIError.Error()
}

// RateLimitError is 429 — too many requests. Retryable after the
// RetryAfterSeconds interval (parsed from the Retry-After header).
type RateLimitError struct {
	APIError
	RetryAfterSeconds float64
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("socialapis: rate-limited, retry after %.0fs: %s", e.RetryAfterSeconds, e.APIError.Error())
}

// BadRequestError is 4xx (excluding 401/402/429) — client-side mistake.
// NOT retryable without fixing input.
type BadRequestError struct{ APIError }

func (e *BadRequestError) Error() string {
	return "socialapis: bad request: " + e.APIError.Error()
}

// APIServerError is 5xx — the API failed. Safe to retry with backoff.
type APIServerError struct{ APIError }

func (e *APIServerError) Error() string {
	return "socialapis: server error: " + e.APIError.Error()
}

// ConnectionError is a network failure / timeout / non-JSON response.
// Almost always transient. Safe to retry with backoff.
type ConnectionError struct {
	Message string
	Cause   error
}

func (e *ConnectionError) Error() string {
	if e.Cause != nil {
		return "socialapis: connection error: " + e.Message + ": " + e.Cause.Error()
	}
	return "socialapis: connection error: " + e.Message
}

func (e *ConnectionError) Unwrap() error { return e.Cause }

func (e *ConnectionError) socialapisError() {}
