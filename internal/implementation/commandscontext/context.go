package commandscontext

type CommandContextImpl struct {
	guildID   string
	channelID string
}

func (cc *CommandContextImpl) GetGuildID() string {
	return cc.guildID
}

func (cc *CommandContextImpl) GetChannelID() string {
	return cc.channelID
}
