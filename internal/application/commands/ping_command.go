package commands

import (
	"github.com/marouane-souiri/vocalize/internal/application/commands/commandsmanager"
	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

type PingCommand struct {
	commandsmanager.BaseCommandImpl
}

func NewPingCommand() interfaces.BaseCommand {
	c := &PingCommand{}
	c.Name = "ping"
	c.Description = "A ping command"
	return c
}

func (cmd *PingCommand) Run(c interfaces.Client, ctx interfaces.CommandContext) error {
	c.SendMessage("1250930757254516836", &domain.SendMessage{
		Content: "Pong !!",
	})
	return nil
}
