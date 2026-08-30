package security

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrMessageTooLarge = errors.New("message too large")
	ErrTooManyClients  = errors.New("too many clients")
)

type WebSocketGuard struct {
	mu          sync.Mutex
	clients     int
	maxClients  int
	rateLimiter *RateLimiter
}

func NewWebSocketGuard(maxClients int, requests int) *WebSocketGuard {
	if maxClients < 1 {
		maxClients = 100
	}

	if requests < 1 {
		requests = 20
	}

	return &WebSocketGuard{
		maxClients:  maxClients,
		rateLimiter: NewRateLimiter(requests, time.Second),
	}
}

func (g *WebSocketGuard) Connect(sessionID string) error {
	if err := ValidateSessionID(sessionID); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.clients >= g.maxClients {
		return ErrTooManyClients
	}

	g.clients++
	return nil
}

func (g *WebSocketGuard) Disconnect() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.clients > 0 {
		g.clients--
	}
}

func (g *WebSocketGuard) AllowMessage(sessionID string, size int64) error {
	if err := ValidateSessionID(sessionID); err != nil {
		return err
	}

	if !ValidMessageSize(size) {
		return ErrMessageTooLarge
	}

	if !g.rateLimiter.Allow(sessionID) {
		return ErrTooManyClients
	}

	return nil
}

func (g *WebSocketGuard) ClientCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.clients
}

func (g *WebSocketGuard) AddClient() error {
	return g.Connect(newInternalSessionID())
}

func (g *WebSocketGuard) RemoveClient() {
	g.Disconnect()
}

func (g *WebSocketGuard) CheckMessageSize(size int64) error {
	if !ValidMessageSize(size) {
		return ErrMessageTooLarge
	}

	return nil
}

func (g *WebSocketGuard) Allow(key string) bool {
	if key == "" {
		return false
	}

	return g.rateLimiter.Allow(key)
}

func newInternalSessionID() string {
	return "00000000000000000000000000000001"
}
