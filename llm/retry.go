package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/googleapi"
)

const maxRetries = 3

var retryBackoffs = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	4 * time.Second,
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		switch gerr.Code {
		case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway,
			http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "resource exhausted") ||
		strings.Contains(msg, "unavailable") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "temporarily unavailable")
}

func withRetry(ctx context.Context, fn func(context.Context) (*WineInfoResult, error)) (*WineInfoResult, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled before gemini request: %w", err)
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !isRetryable(err) || attempt == maxRetries-1 {
			break
		}

		delay := retryBackoffs[attempt]
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("gemini request failed after %d attempts: %w", maxRetries, lastErr)
}
