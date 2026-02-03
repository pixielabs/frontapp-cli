package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiter struct {
	mu             sync.Mutex
	limit          int
	remaining      int
	burstLimit     int
	burstRemaining int
	resetAt        time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (r *RateLimiter) UpdateFromHeaders(h http.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.limit = headerInt(h, "x-ratelimit-limit", r.limit)
	r.remaining = headerInt(h, "x-ratelimit-remaining", r.remaining)
	r.burstLimit = headerInt(h, "x-ratelimit-burst-limit", r.burstLimit)
	r.burstRemaining = headerInt(h, "x-ratelimit-burst-remaining", r.burstRemaining)

	if reset := h.Get("x-ratelimit-reset"); reset != "" {
		if ts, err := strconv.ParseInt(reset, 10, 64); err == nil {
			r.resetAt = time.Unix(ts, 0)
		} else if parsed, err := http.ParseTime(reset); err == nil {
			r.resetAt = parsed
		}
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	limit := r.limit
	remaining := r.remaining
	burstRemaining := r.burstRemaining
	resetAt := r.resetAt
	r.mu.Unlock()

	if limit <= 0 || resetAt.IsZero() {
		return nil
	}

	extra := 0

	if burstRemaining > 0 {
		maxExtra := limit / 2
		if burstRemaining < maxExtra {
			extra = burstRemaining
		} else {
			extra = maxExtra
		}
	}

	effectiveRemaining := remaining + extra
	if effectiveRemaining <= 1 {
		return sleepUntil(ctx, resetAt)
	}

	interval := time.Until(resetAt) / time.Duration(effectiveRemaining)
	if interval <= 0 {
		return nil
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("rate limit wait interrupted: %w", ctx.Err())
	}
}

func sleepUntil(ctx context.Context, t time.Time) error {
	d := time.Until(t)
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("rate limit wait interrupted: %w", ctx.Err())
	}
}

func headerInt(h http.Header, key string, fallback int) int {
	v := h.Get(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}
