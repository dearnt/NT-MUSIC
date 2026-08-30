package security

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	limit   int
	window  time.Duration
}

type rateEntry struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}

	if window <= 0 {
		window = time.Second
	}

	return &RateLimiter{
		entries: make(map[string]*rateEntry),
		limit:   limit,
		window:  window,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	if key == "" {
		return false
	}

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.entries[key]

	if !exists || !now.Before(entry.resetTime) {
		r.entries[key] = &rateEntry{
			count:     1,
			resetTime: now.Add(r.window),
		}
		return true
	}

	if entry.count >= r.limit {
		return false
	}

	entry.count++
	return true
}

func (r *RateLimiter) Reset(key string) {
	if key == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.entries, key)
}
