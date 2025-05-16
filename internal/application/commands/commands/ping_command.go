package commands

import (
	"log"

	"github.com/marouane-souiri/vocalize/internal/application/commands/commandsmanager"
	"github.com/marouane-souiri/vocalize/internal/discord/client"
)

type PingCommand struct {
	commandsmanager.BaseCommandImpl
}

func NewPingCommand() commandsmanager.BaseCommand {
	command := &PingCommand{}
	command.Name = "ping"
	command.Description = "A ping command"
	return command
}

func (cmd *PingCommand) Run(c client.Client, ctx commandsmanager.CommandContext) error {
	log.Print("Pong !!")
	return nil
}
