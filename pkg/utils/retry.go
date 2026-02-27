package utils

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/jackc/pgconn"
)

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// IsRetryableError checks if an error is retryable (wrapped in RetryableError or is a transient DB error).
// Transient errors include:
//   - RetryableError wrapper (explicitly marked as retryable)
//   - PostgreSQL deadlock (40P01)
//   - PostgreSQL serialization failure (40001)
//   - PostgreSQL connection errors (08006, 08003)
func IsRetryableError(err error) bool {
	// Check if explicitly wrapped as retryable
	var retryableErr *RetryableError
	if errors.As(err, &retryableErr) {
		return true
	}

	// Check for transient PostgreSQL errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40P01": // deadlock_detected
			return true
		case "40001": // serialization_failure
			return true
		case "08006": // connection_failure
			return true
		case "08003": // connection_does_not_exist
			return true
		default:
			return false
		}
	}

	return false
}

// Retry executes fn up to 'attempts' times, with exponential backoff and jitter.
// Only retries if the error is marked as retryable via RetryableError or is a transient DB error.
// sleepMs is the base sleep duration in milliseconds.
// Returns the result and nil error if fn succeeds, or zero value of T and last error if all attempts fail.
func Retry[T any](attempts int, sleepMs int, fn func() (T, error)) (T, error) {
	var lastErr error
	var Zero T

	for attempt := 0; attempt < attempts; attempt++ {
		value, err := fn()
		if err == nil {
			return value, nil // success
		}
		lastErr = err

		// Only retry if error is retryable AND we have more attempts left
		if attempt < attempts-1 && IsRetryableError(err) {
			// Calculate exponential backoff and use jitter to add randomness
			// Formula for exponential backoff: baseDelay * 2^attempt
			exponentialDelay := time.Duration(sleepMs) * time.Millisecond * time.Duration(math.Pow(2, float64(attempt)))
			jitter := time.Duration(rand.Intn(int(exponentialDelay)))
			totalDelay := exponentialDelay + jitter

			time.Sleep(totalDelay)
		} else if !IsRetryableError(err) {
			// Non-retryable error, stop immediately
			return Zero, err
		}
	}

	return Zero, lastErr
}
