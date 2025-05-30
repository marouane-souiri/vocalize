package handlers

import (
	"encoding/json"
	"log"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#guild-member-update
func MemberUpdateHandler(c interfaces.Client) domain.ClientHandler {
	return func(event json.RawMessage) {
		var guildMemberUpdate domain.GuildMemberUpdateEvent

		if err := json.Unmarshal(event, &guildMemberUpdate); err != nil {
			log.Printf("[Handlers] Error unmarshaling GUILD_MEMBER_UPDATE event: %v", err)
			return
		}

		oldMember, err := c.GetMember(guildMemberUpdate.ID, guildMemberUpdate.GuildID)
		if err != nil {
			log.Printf("[Handlers] Error in GUILD_MEMBER_UPDATE event: %v", err)
		}
		newMember := *oldMember

		newMember.Avatar = guildMemberUpdate.Avatar
		newMember.Banner = guildMemberUpdate.Banner
		newMember.Flags = guildMemberUpdate.Flags
		newMember.Nickname = guildMemberUpdate.Nickname
		newMember.Mute = guildMemberUpdate.Mute
		newMember.Deaf = guildMemberUpdate.Deaf
		newMember.JoinedAt = guildMemberUpdate.JoinedAt
		newMember.Roles = guildMemberUpdate.Roles

		c.SetMember(&newMember)
	}
}
