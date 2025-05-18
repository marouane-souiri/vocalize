package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#guild-create
func GuildCreateHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var guildCreate domain.GuildCreateEvent

		if err := json.Unmarshal(event, &guildCreate); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_CREATE event: %v", err)
			return
		}

		if guildCreate.Unavailable {
			return
		}

		c.SetGuild(&guildCreate.Guild)
	}
}
