package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#channel-delete
func ChannelDeleteHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var channelDelete domain.ChannelDeleteEvent

		if err := json.Unmarshal(event, &channelDelete); err != nil {
			log.Printf("[Handlers] Error unmarshaling CHANNEL_DELETE event: %v", err)
			return
		}

		c.DelChannel(channelDelete.ID)
	}
}
