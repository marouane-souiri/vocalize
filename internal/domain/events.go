package domain

import (
	"time"
)

type ReadyEvent struct {
	SessionID        string `json:"session_id"`
	ResumeGatewayURL string `json:"resume_gateway_url"`
	User             struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Discriminator string `json:"discriminator"`
	} `json:"user"`
	Guilds []UnavailableGuild `json:"guilds"`
}

type GuildCreateEvent struct {
	Guild
	JoinedAt time.Time `json:"joined_at"`
}

type GuildUpdateEvent struct {
	Guild
}

type GuildDeleteEvent struct {
	UnavailableGuild
}

type MessageCreateEvent struct {
	GuildID *string `json:"guild_id"`
	Member  *Member `json:"member"`
	Message
}

type ChannelCreateEvent struct {
	Channel
}

type ChannelUpdateEvent struct {
	Channel
}

type ChannelDeleteEvent struct {
	Channel
}

type GuildMemberAddEvent struct {
	Member
}

type GuildMemberRemoveEvent struct {
	GuildID string `json:"guild_id"`
	User
}

type GuildMemberUpdateEvent struct {
	ID       string      `json:"id"`
	GuildID  string      `json:"guild_id"`
	Nickname string      `json:"nick"`
	Avatar   string      `json:"avatar"`
	Banner   string      `json:"banner"`
	Roles    []string    `json:"roles"`
	JoinedAt time.Time   `json:"joined_at"`
	Flags    MemberFlags `json:"flags"`
	Deaf     bool        `json:"deaf"`
	Mute     bool        `json:"mute"`
}
