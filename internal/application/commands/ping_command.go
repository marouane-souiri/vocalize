package commands

import (
	"github.com/marouane-souiri/vocalize/internal/application/commands/commandsmanager"
	"github.com/marouane-souiri/vocalize/internal/discord/client"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

type PingCommand struct {
	commandsmanager.BaseCommandImpl
}

func NewPingCommand() commandsmanager.BaseCommand {
	c := &PingCommand{}
	c.Name = "ping"
	c.Description = "A ping command"
	return c
}

func (cmd *PingCommand) Run(c client.Client, ctx commandsmanager.CommandContext) error {
	c.SendMessage("1250930757254516836", &models.SendMessage{
		Content: "Pong !!",
	})
	return nil
}
