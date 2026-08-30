package security

import (
	"errors"
	"time"
)

var ErrActionBlocked = errors.New("action temporarily blocked")

type AbuseGuard struct {
	joinLimiter    *RateLimiter
	actionLimiter  *RateLimiter
	messageLimiter *RateLimiter
}

func NewAbuseGuard() *AbuseGuard {
	return &AbuseGuard{
		joinLimiter:    NewRateLimiter(5, time.Minute),
		actionLimiter:  NewRateLimiter(30, time.Minute),
		messageLimiter: NewRateLimiter(60, time.Minute),
	}
}

func (g *AbuseGuard) CheckJoin(key string) error {
	if key == "" || !g.joinLimiter.Allow(key) {
		return ErrActionBlocked
	}

	return nil
}

func (g *AbuseGuard) CheckAction(key string) error {
	if key == "" || !g.actionLimiter.Allow(key) {
		return ErrActionBlocked
	}

	return nil
}

func (g *AbuseGuard) CheckMessage(key string) error {
	if key == "" || !g.messageLimiter.Allow(key) {
		return ErrActionBlocked
	}

	return nil
}
