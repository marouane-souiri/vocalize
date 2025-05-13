package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	gatewayURL       = "wss://gateway.discord.gg/?v=10&encoding=json"
	opDispatch       = 0
	opHeartbeat      = 1
	opIdentify       = 2
	opResume         = 6
	opReconnect      = 7
	opInvalidSession = 9
	opHello          = 10
	opHeartbeatACK   = 11
)

type Client interface {
	Start() error
	Stop() error
	On(eventType string, handler func(event json.RawMessage))
	Once(eventType string, handler func(event json.RawMessage))
}

type Payload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
	S  int             `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
}

type clientHandlerKind int

const (
	clientHandlerNormal       = -1
	clientHandlerOnce         = 1
	clientHandlerConsumedOnce = 0
)

type clientHandler struct {
	run  func(event json.RawMessage)
	kind clientHandlerKind
}

type clientImpl struct {
	token string
	ws    WSManager

	sessionID        string
	resumeGatewayURL string
	sequence         int

	heartbeatInterval time.Duration
	heartbeatCancel   context.CancelFunc
	lastHeartbeatAck  time.Time

	eventHandlers map[string][]*clientHandler
	mu            sync.RWMutex

	shutdown     chan struct{}
	reconnecting bool
	reconnectMu  sync.Mutex
}

func NewClient(wsManager WSManager, token string) (Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	return &clientImpl{
		token:         token,
		ws:            wsManager,
		eventHandlers: make(map[string][]*clientHandler),
		shutdown:      make(chan struct{}),
	}, nil
}

func (c *clientImpl) Start() error {
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

func (c *clientImpl) On(eventType string, handler func(event json.RawMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers[eventType] = append(c.eventHandlers[eventType], &clientHandler{
		run:  handler,
		kind: clientHandlerNormal,
	})
}

func (c *clientImpl) Once(eventType string, handler func(event json.RawMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers[eventType] = append(c.eventHandlers[eventType], &clientHandler{
		run:  handler,
		kind: clientHandlerOnce,
	})
}

func (c *clientImpl) listenForEvents() {
	for {
		select {
		case <-c.shutdown:
			return
		case message := <-c.ws.Receive():
			// TODO: Run handleMessage using a workerpool
			c.handleMessage(message)
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
		reconnectURL = gatewayURL
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

func (c *clientImpl) handleDispatch(eventType string, data json.RawMessage) {
	log.Printf("[Discord] Processing event: %s", eventType)

	if eventType == "READY" {
		c.handleReady(data)
	} else if eventType == "RESUMED" {
		log.Println("[Discord] Session resumed successfully")
	}

	c.mu.RLock()
	handlers, exists := c.eventHandlers[eventType]
	c.mu.RUnlock()

	if exists {
		for i, handler := range handlers {
			if handler.kind == clientHandlerConsumedOnce {
				continue
			}

			if handler.kind == clientHandlerOnce {
				c.mu.Lock()
				if i < len(c.eventHandlers[eventType]) {
					c.eventHandlers[eventType][i].kind = clientHandlerConsumedOnce
				}
				c.mu.Unlock()
			}

			handler.run(data)
		}
	}
}

func (c *clientImpl) handleReady(data json.RawMessage) {
	var ready struct {
		SessionID        string `json:"session_id"`
		ResumeGatewayURL string `json:"resume_gateway_url"`
		User             struct {
			ID            string `json:"id"`
			Username      string `json:"username"`
			Discriminator string `json:"discriminator"`
		} `json:"user"`
	}
	if err := json.Unmarshal(data, &ready); err != nil {
		log.Printf("[Discord] Error unmarshaling READY event: %v", err)
		return
	}

	c.sessionID = ready.SessionID
	c.resumeGatewayURL = ready.ResumeGatewayURL
	if c.resumeGatewayURL == "" {
		c.resumeGatewayURL = gatewayURL
	}

	log.Printf("[Discord] Connected as %s#%s", ready.User.Username, ready.User.Discriminator)
	log.Printf("[Discord] Session ID: %s", c.sessionID)
	log.Printf("[Discord] Resume Gateway URL: %s", c.resumeGatewayURL)
}

func (c *clientImpl) handleInvalidSession(data json.RawMessage) {
	var canResume bool
	if err := json.Unmarshal(data, &canResume); err != nil {
		log.Printf("[Discord] Error unmarshaling invalid session data: %v", err)
		canResume = false
	}

	if canResume {
		log.Println("[Discord] Session is resumable, reconnecting")
		c.handleReconnect()
	} else {
		log.Println("[Discord] Session not resumable, creating new session")

		c.sessionID = ""
		c.sequence = 0

		waitTime := time.Duration(rand.Intn(4000)+1000) * time.Millisecond
		log.Printf("[Discord] Waiting %v before identifying", waitTime)
		time.Sleep(waitTime)

		c.sendIdentify()
	}
}
