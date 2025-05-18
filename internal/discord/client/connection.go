package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

func (c *clientImpl) Start() error {
	c.ws.SetUrl(c.url)
	go c.listenForEvents()
	return c.ws.Connect()
}

func (c *clientImpl) Stop() error {
	close(c.shutdown)
	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}
	return c.ws.Close()
}

func (c *clientImpl) listenForEvents() {
	for {
		select {
		case <-c.shutdown:
			return
		case message := <-c.ws.Receive():
			c.wp.Submit(func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("[Client] Recovered from a message handler")
					}
				}()
				c.handleMessage(message)
			})
		case err := <-c.ws.Errors():
			log.Printf("[WebSocket] Error: %v", err)
			if shouldReconnect(err) {
				c.handleReconnect()
			}
		}
	}
}

func shouldReconnect(err error) bool {
	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}

	if strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}

	if strings.Contains(err.Error(), "EOF") {
		return true
	}

	return false
}

func (c *clientImpl) handleMessage(message []byte) {
	var payload Payload
	if err := json.Unmarshal(message, &payload); err != nil {
		log.Printf("[Discord] Error unmarshaling payload: %v", err)
		return
	}

	if payload.S != 0 {
		c.sequence = payload.S
	}

	switch payload.Op {
	case opHello:
		log.Println("[Discord] Received hello event")
		c.handleHello(payload.D)
	case opHeartbeatACK:
		log.Println("[Discord] Received heartbeat ACK")
		c.lastHeartbeatAck = time.Now()
	case opHeartbeat:
		log.Println("[Discord] Received heartbeat req")
		c.sendHeartbeat()
	case opReconnect:
		log.Println("[Discord] Received reconnect request")
		c.handleReconnect()
	case opInvalidSession:
		log.Println("[Discord] Invalid session")
		c.handleInvalidSession(payload.D)
	case opDispatch:
		log.Println("[Discord] Received dispatch event:", payload.T)
		c.handleDispatch(payload.T, payload.D)
	}
}

func (c *clientImpl) handleReconnect() {
	c.reconnectMu.Lock()
	if c.reconnecting {
		c.reconnectMu.Unlock()
		return
	}
	c.reconnecting = true
	c.reconnectMu.Unlock()

	c.authMu.Lock()
	c.authenticated = false
	c.authMu.Unlock()

	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
		c.heartbeatCancel = nil
	}

	maxRetries := 5
	baseDelay := 1 * time.Second

	var reconnectURL string
	if c.resumeGatewayURL != "" {
		reconnectURL = c.resumeGatewayURL
	} else {
		reconnectURL = c.url
	}

	var success bool

	for attempt := 0; attempt < maxRetries && !success; attempt++ {
		delay := baseDelay * time.Duration(1<<attempt) // 1s, 2s, 4s, 8s, 16s
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.3)
		reconnectDelay := delay + jitter

		log.Printf("[Discord] Reconnection attempt %d/%d, waiting %v", attempt+1, maxRetries, reconnectDelay)
		time.Sleep(reconnectDelay)

		if err := c.ws.Reconnect(reconnectURL); err != nil {
			log.Printf("[Discord] Failed to reconnect: %v, will retry", err)
			continue
		}

		log.Println("[Discord] Reconnected, waiting for Hello")
		success = true
	}

	if !success {
		log.Printf("[Discord] Failed to reconnect after %d attempts, trying fresh connection", maxRetries)
		time.Sleep(5 * time.Second)

		if err := c.ws.Connect(); err != nil {
			log.Printf("[Discord] Fresh connection also failed: %v", err)
		}
	}

	c.reconnectMu.Lock()
	c.reconnecting = false
	c.reconnectMu.Unlock()
}

func (c *clientImpl) startHeartbeat() {
	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.heartbeatCancel = cancel

	minInterval := time.Duration(float64(c.heartbeatInterval) * 0.9)
	maxInterval := c.heartbeatInterval
	interval := minInterval + time.Duration(rand.Float64()*float64(maxInterval-minInterval))

	ticker := time.NewTicker(interval)
	c.lastHeartbeatAck = time.Now()

	go func() {
		initialJitter := time.Duration(rand.Float64() * float64(interval))
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-time.After(initialJitter):
		}

		c.sendHeartbeat()

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if time.Since(c.lastHeartbeatAck) > c.heartbeatInterval*2 && !c.lastHeartbeatAck.IsZero() {
					log.Println("[Discord] No heartbeat ACK received, reconnecting")
					ticker.Stop()
					go c.handleReconnect()
					return
				}
				c.sendHeartbeat()
			}
		}
	}()
}

func (c *clientImpl) sendHeartbeat() {
	payload := Payload{
		Op: opHeartbeat,
		D:  json.RawMessage(fmt.Sprintf("%d", c.sequence)),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Discord] Error marshaling heartbeat: %v", err)
		return
	}

	log.Println("[Discord] Sending heartbeat")
	c.ws.Send(data)
}

func (c *clientImpl) sendIdentify() {
	c.authMu.Lock()
	if c.authenticated {
		log.Println("[Discord] Already authenticated, skipping identify")
		c.authMu.Unlock()
		return
	}
	c.authMu.Unlock()

	identify := Payload{
		Op: opIdentify,
		D: json.RawMessage(fmt.Sprintf(`{
			"token": "%s",
			"properties": {
				"os": "linux",
				"browser": "discord-client",
				"device": "discord-client"
			},
			"intents": %s
		}`, c.token, c.intents)),
	}

	data, err := json.Marshal(identify)
	if err != nil {
		log.Printf("[Discord] Error marshaling identify: %v", err)
		return
	}

	log.Println("[Discord] Sending identify payload")
	c.ws.Send(data)
}

func (c *clientImpl) sendResume() {
	c.authMu.Lock()
	c.authenticated = false
	c.authMu.Unlock()

	resume := Payload{
		Op: opResume,
		D: json.RawMessage(fmt.Sprintf(`{
			"token": "%s",
			"session_id": "%s",
			"seq": %d
		}`, c.token, c.sessionID, c.sequence)),
	}

	data, err := json.Marshal(resume)
	if err != nil {
		log.Printf("[Discord] Error marshaling resume: %v", err)
		return
	}

	log.Println("[Discord] Sending resume payload")
	c.ws.Send(data)
}
