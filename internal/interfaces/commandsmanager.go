package interfaces

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
