package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 1024 * 1024
)

type WSClient struct {
	mu        sync.RWMutex
	writeMu   sync.Mutex
	conn      *websocket.Conn
	url       string
	connected bool
}

func NewWSClient(url string) *WSClient {
	return &WSClient{
		url: url,
	}
}

func (c *WSClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.connected {
		c.mu.Unlock()
		return errors.New("already connected")
	}
	c.mu.Unlock()

	dialer := websocket.DefaultDialer

	conn, _, err := dialer.DialContext(ctx, c.url, http.Header{})
	if err != nil {
		return err
	}

	conn.SetReadLimit(maxMessageSize)

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	return nil
}

func (c *WSClient) Send(message ClientMessage) error {
	c.mu.RLock()
	conn := c.conn
	connected := c.connected
	c.mu.RUnlock()

	if !connected || conn == nil {
		return errors.New("not connected")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	return conn.WriteJSON(message)
}

func (c *WSClient) Read(ctx context.Context) (ServerMessage, error) {
	c.mu.RLock()
	conn := c.conn
	connected := c.connected
	c.mu.RUnlock()

	if !connected || conn == nil {
		return ServerMessage{}, errors.New("not connected")
	}

	deadline, ok := ctx.Deadline()
	if ok {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return ServerMessage{}, err
		}
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		c.markDisconnected()
		return ServerMessage{}, err
	}

	var message ServerMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return ServerMessage{}, err
	}

	return message, nil
}

func (c *WSClient) StartPing(ctx context.Context) error {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			connected := c.connected
			c.mu.RUnlock()

			if !connected || conn == nil {
				return errors.New("not connected")
			}

			c.writeMu.Lock()
			err := conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err == nil {
				err = conn.WriteMessage(websocket.PingMessage, nil)
			}
			c.writeMu.Unlock()

			if err != nil {
				c.markDisconnected()
				return err
			}
		}
	}
}

func (c *WSClient) Close() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.connected = false
	c.mu.Unlock()

	if conn == nil {
		return nil
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return conn.Close()
}

func (c *WSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected && c.conn != nil
}

func (c *WSClient) markDisconnected() {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
}
