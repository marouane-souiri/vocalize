package client

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

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
		Guilds []models.UnavailableGuild `json:"guilds"`
	}
	if err := json.Unmarshal(data, &ready); err != nil {
		log.Printf("[Discord] Error unmarshaling READY event: %v", err)
		return
	}

	c.sessionID = ready.SessionID
	c.resumeGatewayURL = ready.ResumeGatewayURL
	if c.resumeGatewayURL == "" {
		log.Println("[Discord] No Resume Gateway URL provided, using default")
		c.resumeGatewayURL = c.url
	}

	log.Printf("[Discord] Connected as %s#%s", ready.User.Username, ready.User.Discriminator)
	log.Printf("[Discord] Session ID: %s", c.sessionID)
	log.Printf("[Discord] Resume Gateway URL: %s", c.resumeGatewayURL)

	if len(ready.Guilds) == 0 {
		log.Println("[Discord] Info: No Guild to cache")
	} else {
		log.Printf("[Discord] Info: start Caching %d UnavailableGuild", len(ready.Guilds))
		for _, guild := range ready.Guilds {
			if ok := c.cm.SetGuild(&models.Guild{ID: guild.ID, Unavailable: true}); !ok {
				log.Printf("[Discord] Info: guild with id %s, did not get cached", guild.ID)
			}
		}
		c.cm.FlushGuilds()
		log.Println("[Discord] Info: Caching UnavailableGuild ends")
	}
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
