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
