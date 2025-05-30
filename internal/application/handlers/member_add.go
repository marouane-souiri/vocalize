package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#guild-member-add
func MemberAddHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var guildMemberAdd domain.GuildMemberAddEvent

		if err := json.Unmarshal(event, &guildMemberAdd); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_MEMBER_ADD event: %v", err)
			return
		}

		c.SetMember(&guildMemberAdd.Member)
	}
}
