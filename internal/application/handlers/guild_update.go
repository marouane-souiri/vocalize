package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#guild-update
func GuildUpdateHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var guildUpdate domain.GuildUpdateEvent

		if err := json.Unmarshal(event, &guildUpdate); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_UPDATE event: %v", err)
			return
		}

		c.SetGuild(&guildUpdate.Guild)
	}
}
