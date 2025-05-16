package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/discord/client"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

// https://discord.com/developers/docs/events/gateway-events#guild-delete
func GuildDeleteHandler(c client.Client) client.HandlerFunc {
	return func(event json.RawMessage) {
		var guildDelete models.GuildDeleteEvent

		if err := json.Unmarshal(event, &guildDelete); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_DELETE event: %v", err)
			return
		}

		if guildDelete.Unavailable {
			c.SetGuild(&models.Guild{ID: guildDelete.ID, Unavailable: true})
		} else {
			c.DelGuild(guildDelete.ID)
		}
	}
}
