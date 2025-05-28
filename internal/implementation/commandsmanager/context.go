package commandsmanager

import (
	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

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

type CommandsContextMakerImpl struct{}

func NewCommandsContextMaker() interfaces.CommandsContextMaker {
	return &CommandsContextMakerImpl{}
}

func (ccm *CommandsContextMakerImpl) FromMessageEvent(event domain.MessageCreateEvent) interfaces.CommandContext {
	return &CommandContextImpl{
		guildID:   *event.GuildID,
		channelID: event.ChannelID,
	}
}
