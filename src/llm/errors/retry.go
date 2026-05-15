package errors

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

var rateLimitBackoff = []time.Duration{
	1 * time.Second,
	4 * time.Second,
	16 * time.Second,
}

type QuotaDetector interface {
	IsQuotaError(*http.Response, []byte) bool
}

func RetryWithBackoff(detector QuotaDetector, fn func() (*http.Response, error)) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		resp, err = fn()
		if err != nil {
			return resp, err
		}

		if !isRetryableRateLimit(detector, resp) {
			return resp, nil
		}

		if attempt >= len(rateLimitBackoff) {
			return resp, nil
		}

		drainAndClose(resp)
		time.Sleep(rateLimitBackoff[attempt])
	}
}

func isRetryableRateLimit(detector QuotaDetector, resp *http.Response) bool {
	if resp.StatusCode != http.StatusTooManyRequests {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// 429 may be a quota/billing failure rather than a transient rate limit;
	// quota is not retryable and providers detect their own response shape.
	return !detector.IsQuotaError(resp, body)
}

func drainAndClose(resp *http.Response) {
	if resp.Body == nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
