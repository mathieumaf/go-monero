package rpc

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RetryPolicy decides whether a failed attempt of an idempotent call should
// be tried again. attempt is zero-based: it is 0 for the first failure, 1
// for the second, and so on. delay is how long to wait before the next
// attempt. Returning retry == false stops and preserves the last error.
//
// The policy runs synchronously on the calling goroutine and must be safe
// for concurrent use of the same client. It must eventually return false;
// the client does not impose its own attempt cap.
type RetryPolicy func(attempt int, err error) (delay time.Duration, retry bool)

// WithRetry installs an opt-in retry policy for [Client.JSONRPC] and
// [Client.RawRequest]. The default client does not retry: each call performs
// a single attempt, which is the historical behavior.
//
// The policy applies to every call made with the client, including calls that
// are not idempotent. Use it on a client dedicated to idempotent reads, or
// return false from the policy for failures that must not be repeated.
// Context cancellation and deadline expiry are never retried.
//
// Passing a nil policy leaves retries disabled.
func WithRetry(policy RetryPolicy) func(o *clientOptions) {
	return func(o *clientOptions) {
		o.Retry = policy
	}
}

// withRetry runs fn once when no policy is configured. Otherwise it retries
// according to the policy until the call succeeds, the policy stops, or the
// context is done.
func (c *Client) withRetry(ctx context.Context, fn func() error) error {
	if c.retry == nil {
		return fn()
	}

	var attempt int

	for {
		err := fn()
		if err == nil || !canRetry(ctx, err) {
			return err
		}

		delay, retry := c.retry(attempt, err)
		if !retry {
			return err
		}

		if waitErr := wait(ctx, delay); waitErr != nil {
			return fmt.Errorf("retry canceled: %w: %w", waitErr, err)
		}

		attempt++
	}
}

// canRetry reports whether another attempt is allowed. Context errors are
// terminal so a policy cannot loop on a canceled call.
func canRetry(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	return true
}

// wait blocks until delay elapses or ctx ends. A non-positive delay returns
// immediately when the context is still active.
func wait(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
