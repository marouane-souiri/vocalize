package commandsmanager

import "github.com/marouane-souiri/vocalize/internal/discord/client"

type BaseCommand interface {
	GetName() string
	GetAliases() []string
	GetDescription() string
	Run(client client.Client, ctx CommandContext) error
}

type BaseCommandImpl struct {
	Name        string
	Aliases     []string
	Description string
}

func (b *BaseCommandImpl) GetName() string {
	return b.Name
}

func (b *BaseCommandImpl) GetAliases() []string {
	return b.Aliases
}

func (b *BaseCommandImpl) GetDescription() string {
	return b.Description
}
