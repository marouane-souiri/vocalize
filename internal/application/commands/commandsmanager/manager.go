package commandsmanager

import (
	"sync"

	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

type CommandsManagerImpl struct {
	commands map[string]interfaces.BaseCommand
	mu       sync.RWMutex
}

func NewCommandsManager() interfaces.CommandsManager {
	return &CommandsManagerImpl{
		commands: make(map[string]interfaces.BaseCommand),
	}
}

func (c *CommandsManagerImpl) AddCommand(command interfaces.BaseCommand) {
	c.mu.Lock()
	c.commands[command.GetName()] = command
	for _, alias := range command.GetAliases() {
		c.commands[alias] = command
	}
	c.mu.Unlock()
}

func (c *CommandsManagerImpl) AddCommands(commands ...interfaces.BaseCommand) {
	c.mu.Lock()
	for _, command := range commands {
		c.commands[command.GetName()] = command
		for _, alias := range command.GetAliases() {
			c.commands[alias] = command
		}
	}
	c.mu.Unlock()
}

func (c *CommandsManagerImpl) GetCommand(NameOrAlias string) (interfaces.BaseCommand, bool) {
	c.mu.RLock()
	command, ok := c.commands[NameOrAlias]
	c.mu.RUnlock()
	return command, ok
}
