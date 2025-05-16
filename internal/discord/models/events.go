package models

import "time"

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
