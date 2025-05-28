package domain

type ChannelType int

const (
	ChannelType_GUILD_TEXT ChannelType = iota
	ChannelType_DM
	ChannelType_GUILD_VOICE
	ChannelType_GROUP_DM
	ChannelType_GUILD_CATEGORY
	ChannelType_GUILD_ANNOUNCEMENT
	ChannelType_ANNOUNCEMENT_THREAD
	ChannelType_PUBLIC_THREAD
	ChannelType_PRIVATE_THREAD
	ChannelType_GUILD_STAGE_VOICE
	ChannelType_GUILD_DIRECTORY
	ChannelType_GUILD_FORUM
	ChannelType_GUILD_MEDIA
)

type Channel struct {
	ID   string      `json:"id"`
	Type ChannelType `json:"type"`
	// the real channel object
	Data any
}

type GuildTextChannel struct {
}

type DMChannel struct {
}

type GuildAnnouncementChannel struct {
}

type GuildVoiceChannel struct {
}

type GuildCategoryChannel struct {
}
