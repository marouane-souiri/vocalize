package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/discord/client"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

// https://discord.com/developers/docs/events/gateway-events#guild-create
func GuildCreateHandler(c client.Client) client.HandlerFunc {
	return func(event json.RawMessage) {
		var guildCreate models.GuildCreateEvent

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
