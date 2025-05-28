package interfaces

import "github.com/marouane-souiri/vocalize/internal/domain"

type CommandContext interface {
	GetGuildID() string
	GetChannelID() string
}

type CommandsContextMaker interface {
	FromMessageEvent(event domain.MessageCreateEvent) CommandContext
}
