package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#channel-create
func ChannelCreateHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var channelCreate domain.ChannelCreateEvent

		if err := json.Unmarshal(event, &channelCreate); err != nil {
			log.Printf("[Handlers] Error unmarshaling CHANNEL_CREATE event: %v", err)
			return
		}

		switch channelCreate.Type {
		case domain.ChannelType_GUILD_TEXT:
			var guildtext domain.GuildTextChannel
			if err := json.Unmarshal(event, &guildtext); err != nil {
				log.Printf("[Handlers] Error unmarshaling Channel of type GUILD_TEXT: %v", err)
				return
			}
			channelCreate.Channel.Data = guildtext
		case domain.ChannelType_GUILD_VOICE:
			var guildvoice domain.GuildVoiceChannel
			if err := json.Unmarshal(event, &guildvoice); err != nil {
				log.Printf("[Handlers] Error unmarshaling Channel of type GUILD_VOICE: %v", err)
				return
			}
			channelCreate.Channel.Data = guildvoice
		case domain.ChannelType_GUILD_CATEGORY:
			var guildcategory domain.GuildCategoryChannel
			if err := json.Unmarshal(event, &guildcategory); err != nil {
				log.Printf("[Handlers] Error unmarshaling Channel of type GUILD_CATEGORY: %v", err)
				return
			}
			channelCreate.Channel.Data = guildcategory
		default:
			return
		}

		c.SetChannel(&channelCreate.Channel)
	}
}
