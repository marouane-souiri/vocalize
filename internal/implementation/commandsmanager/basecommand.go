package commandsmanager

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
