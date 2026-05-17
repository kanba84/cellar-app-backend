package llm

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"google.golang.org/api/googleapi"
)

func TestIsRetryable(t *testing.T) {
	if !isRetryable(&googleapi.Error{Code: http.StatusTooManyRequests}) {
		t.Fatal("expected 429 to be retryable")
	}
	if isRetryable(context.Canceled) {
		t.Fatal("expected context.Canceled to not be retryable")
	}
	if !isRetryable(errors.New("Error 429: rate limit")) {
		t.Fatal("expected rate limit message to be retryable")
	}
}

func TestWithRetryRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := withRetry(ctx, func(context.Context) (*WineInfoResult, error) {
		return nil, errors.New("should not run")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithRetryStopsAfterMaxAttempts(t *testing.T) {
	attempts := 0
	_, err := withRetry(context.Background(), func(context.Context) (*WineInfoResult, error) {
		attempts++
		return nil, &googleapi.Error{Code: http.StatusTooManyRequests}
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != maxRetries {
		t.Fatalf("attempts = %d, want %d", attempts, maxRetries)
	}
}

func TestWithRetrySucceedsOnSecondAttempt(t *testing.T) {
	start := time.Now()
	attempts := 0

	result, err := withRetry(context.Background(), func(context.Context) (*WineInfoResult, error) {
		attempts++
		if attempts == 1 {
			return nil, &googleapi.Error{Code: http.StatusTooManyRequests}
		}
		return &WineInfoResult{}, nil
	})
	if err != nil {
		t.Fatalf("withRetry() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if time.Since(start) < retryBackoffs[0] {
		t.Fatalf("expected backoff delay, got %v", time.Since(start))
	}
}
