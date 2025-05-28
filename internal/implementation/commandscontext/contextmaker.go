package commandscontext

import (
	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

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
