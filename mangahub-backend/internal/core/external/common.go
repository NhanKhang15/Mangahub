package external

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// rateLimiter is a tiny mutex-guarded token bucket sized to N requests per
// second. It avoids pulling in golang.org/x/time/rate while still giving the
// upstream APIs the politeness window the Phase 3 plan asks for.
type rateLimiter struct {
	interval time.Duration
	mu       sync.Mutex
	next     time.Time
}

func newRateLimiter(rps int) *rateLimiter {
	if rps <= 0 {
		rps = 5
	}
	return &rateLimiter{interval: time.Second / time.Duration(rps)}
}

func (l *rateLimiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	now := time.Now()
	var sleep time.Duration
	if !l.next.IsZero() && now.Before(l.next) {
		sleep = l.next.Sub(now)
	}
	l.next = now.Add(sleep + l.interval)
	l.mu.Unlock()

	if sleep <= 0 {
		return nil
	}
	t := time.NewTimer(sleep)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// doWithRetry executes the request built by `build` against `client`, applying
// the rate limiter and retrying up to `maxAttempts` times on HTTP 429 / 5xx
// with exponential backoff. Retry-After (seconds or HTTP-date) is honored.
//
// The response body is read fully, the response is closed, and the bytes are
// returned alongside the *http.Response (whose Body is already drained).
func doWithRetry(
	ctx context.Context,
	client *http.Client,
	build func() (*http.Request, error),
	limiter *rateLimiter,
	maxAttempts int,
) (*http.Response, []byte, error) {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	backoff := 500 * time.Millisecond

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if limiter != nil {
			if err := limiter.Wait(ctx); err != nil {
				return nil, nil, err
			}
		}
		req, err := build()
		if err != nil {
			return nil, nil, err
		}
		req = req.WithContext(ctx)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			switch {
			case readErr != nil:
				lastErr = readErr
			case resp.StatusCode < 400:
				return resp, body, nil
			case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
				lastErr = fmt.Errorf("upstream %d: %s", resp.StatusCode, truncate(body, 200))
				if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
					backoff = ra
				}
			default:
				// 4xx other than 429 — don't retry, bubble up.
				return resp, body, fmt.Errorf("upstream %d: %s", resp.StatusCode, truncate(body, 200))
			}
		}

		if attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return nil, nil, fmt.Errorf("retries exhausted: %w", lastErr)
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	if secs, err := strconv.Atoi(h); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
