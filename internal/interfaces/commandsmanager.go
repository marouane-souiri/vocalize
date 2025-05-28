package interfaces

import "github.com/marouane-souiri/vocalize/internal/domain"

type CommandContext interface {
	GetGuildID() string
	GetChannelID() string
}

type CommandsContextMaker interface {
	FromMessageEvent(event domain.MessageCreateEvent) CommandContext
}

type BaseCommand interface {
	GetName() string
	GetAliases() []string
	GetDescription() string
	Run(client Client, ctx CommandContext) error
}

type CommandsManager interface {
	AddCommand(command BaseCommand)
	AddCommands(commands ...BaseCommand)
	GetCommand(NameOrAlias string) (BaseCommand, bool)
}
