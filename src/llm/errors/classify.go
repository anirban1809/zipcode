package errors

import (
	"fmt"
	"io"
	"net/http"
)

type ErrorBucket int

const (
	BucketAuth         ErrorBucket = iota // 401, 403
	BucketQuotaBilling                    // 402, 429 with quota markers, provider-specific
	BucketRateLimit                       // 429 transient
	BucketModelMissing                    // 404 with model_not_found, similar
	BucketNetwork                         // network errors, timeouts, 5xx
	BucketUnspecified
)

type ClassifiedError struct {
	Bucket      ErrorBucket
	Provider    string
	UserMessage string // plain-language, includes the next-step hint
	RawDetails  string // post-redaction copy of the raw error for the collapsed Show details
	Retryable   bool
}

var messageTemplates map[ErrorBucket]string = map[ErrorBucket]string{
	BucketAuth:         "%s rejected this key. It may be revoked, expired, or pasted incorrectly. Run /providers to update it.",
	BucketModelMissing: "%s doesn't recognize '{model}'. Switch with /models.",
	BucketRateLimit:    "%s is rate-limiting this key. Retried 3 times. Try again in a minute.",
	BucketNetwork:      "Couldn't reach %s (timed out after 30s). Check your connection, then retry.",
	BucketQuotaBilling: "%s reports this key is out of credits or over its spend cap. Add credits, or switch providers with /models.",
	BucketUnspecified:  "Unknown error occurred, please retry or restart the app",
}

func (e ClassifiedError) GetErrorMessage(bucket ErrorBucket) string {
	return fmt.Sprintf(messageTemplates[bucket], e.Provider)
}

func (e ClassifiedError) ClassifyError(provider string, detector QuotaDetector, resp *http.Response, err error) ClassifiedError {
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return ClassifiedError{
			Bucket:      BucketAuth,
			Provider:    provider,
			UserMessage: e.GetErrorMessage(BucketAuth),
			RawDetails:  resp.Status,
			Retryable:   false,
		}
	}

	if resp.StatusCode > 500 || resp.StatusCode == 429 {
		bucket := BucketRateLimit

		retryable := true

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if detector.IsQuotaError(resp, body) {
			bucket = BucketQuotaBilling
			retryable = false
		}

		if resp.StatusCode > 500 {
			bucket = BucketNetwork
		}

		return ClassifiedError{
			Bucket:      bucket,
			Provider:    provider,
			UserMessage: e.GetErrorMessage(bucket),
			RawDetails:  resp.Status,
			Retryable:   retryable,
		}
	}

	if resp.StatusCode == 404 {
		return ClassifiedError{
			Bucket:      BucketModelMissing,
			Provider:    provider,
			UserMessage: e.GetErrorMessage(BucketModelMissing),
			RawDetails:  resp.Status,
			Retryable:   false,
		}
	}

	return ClassifiedError{
		Bucket:      BucketUnspecified,
		Provider:    provider,
		UserMessage: e.GetErrorMessage(BucketUnspecified),
		RawDetails:  resp.Status,
		Retryable:   true,
	}
}
