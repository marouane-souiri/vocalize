package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
			// DONE: Run handleMessage using a workerpool
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
			c.handleReconnect()
		}
	}
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

	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}

	time.Sleep(1 * time.Second)

	var reconnectURL string
	if c.resumeGatewayURL != "" {
		reconnectURL = c.resumeGatewayURL
	} else {
		reconnectURL = c.url
	}

	if err := c.ws.Reconnect(reconnectURL); err != nil {
		log.Printf("[Discord] Failed to reconnect: %v, will make fresh conn", err)

		time.Sleep(3 * time.Second)

		if err := c.ws.Connect(); err != nil {
			// What to do next? reconnect did not work connect did not work too, fuck this just panic.
			panic("Connect and Reconnect both did not work")
		}

		c.reconnectMu.Lock()
		c.reconnecting = false
		c.reconnectMu.Unlock()

		return
	}

	log.Println("[Discord] Reconnected, waiting for Hello")
	c.reconnectMu.Lock()
	c.reconnecting = false
	c.reconnectMu.Unlock()
}

func (c *clientImpl) handleHello(data json.RawMessage) {
	var hello struct {
		HeartbeatInterval int `json:"heartbeat_interval"`
	}
	if err := json.Unmarshal(data, &hello); err != nil {
		log.Printf("[Discord] Error unmarshaling Hello payload: %v", err)
		return
	}

	c.heartbeatInterval = time.Duration(hello.HeartbeatInterval) * time.Millisecond
	log.Printf("[Discord] Received Hello with heartbeat interval: %v", c.heartbeatInterval)

	c.startHeartbeat()

	if c.sessionID != "" && c.sequence > 0 {
		c.sendResume()
	} else {
		c.sendIdentify()
	}
}

func (c *clientImpl) startHeartbeat() {
	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.heartbeatCancel = cancel

	interval := time.Duration(float64(c.heartbeatInterval) * 0.9)
	ticker := time.NewTicker(interval)
	c.lastHeartbeatAck = time.Now()

	go func() {
		jitter := rand.Float64()
		time.Sleep(time.Duration(float64(c.heartbeatInterval) * jitter))

		c.sendHeartbeat()

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if time.Since(c.lastHeartbeatAck) > c.heartbeatInterval*2 && !c.lastHeartbeatAck.IsZero() {
					log.Println("[Discord] No heartbeat ACK received, reconnecting")
					go c.handleReconnect()
					ticker.Stop()
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
	identify := Payload{
		Op: opIdentify,
		D: json.RawMessage(fmt.Sprintf(`{
			"token": "%s",
			"properties": {
				"os": "linux",
				"browser": "discord-client",
				"device": "discord-client"
			},
			"intents": 513
		}`, c.token)),
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
