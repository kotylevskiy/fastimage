package fastimage

import (
	"fmt"
	"io"
	"time"
)

type HTTPStatusError struct {
	URL        string
	StatusCode int
	Status     string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("fastimage: unexpected HTTP status %s", e.Status)
}

type RetryAfterError struct {
	URL        string
	StatusCode int
	Status     string
	RetryAfter time.Duration
}

func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("fastimage: retry after %s", e.Status)
}

type InsufficientBytesError struct {
	URL string
	Got int
	Min int
}

func (e *InsufficientBytesError) Error() string {
	return fmt.Sprintf("fastimage: insufficient bytes: got %d, need at least %d", e.Got, e.Min)
}

// Let callers use errors.Is(err, io.ErrUnexpectedEOF).
func (e *InsufficientBytesError) Unwrap() error { return io.ErrUnexpectedEOF }
