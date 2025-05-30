package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#guild-member-remove
func MemberRemoveHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var guildMemberRemove domain.GuildMemberRemoveEvent

		if err := json.Unmarshal(event, &guildMemberRemove); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_MEMBER_REMOVE event: %v", err)
			return
		}

		c.DelMember(guildMemberRemove.ID, guildMemberRemove.GuildID)
	}
}
