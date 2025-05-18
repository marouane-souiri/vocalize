package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#guild-delete
func GuildDeleteHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var guildDelete domain.GuildDeleteEvent

		if err := json.Unmarshal(event, &guildDelete); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_DELETE event: %v", err)
			return
		}

		if guildDelete.Unavailable {
			c.SetGuild(&domain.Guild{ID: guildDelete.ID, Unavailable: true})
		} else {
			c.DelGuild(guildDelete.ID)
		}
	}
}
