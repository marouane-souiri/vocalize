package commandsmanager

import (
	"sync"
)

type CommandsManager interface {
	AddCommand(command BaseCommand)
	AddCommands(commands ...BaseCommand)
	GetCommand(NameOrAlias string) (BaseCommand, bool)
}

type CommandsManagerImpl struct {
	commands map[string]BaseCommand
	mu       sync.RWMutex
}

func NewCommandsManager() CommandsManager {
	return &CommandsManagerImpl{
		commands: make(map[string]BaseCommand),
	}
}

func (c *CommandsManagerImpl) AddCommand(command BaseCommand) {
	c.mu.Lock()
	c.commands[command.GetName()] = command
	for _, alias := range command.GetAliases() {
		c.commands[alias] = command
	}
	c.mu.Unlock()
}

func (c *CommandsManagerImpl) AddCommands(commands ...BaseCommand) {
	c.mu.Lock()
	for _, command := range commands {
		c.commands[command.GetName()] = command
		for _, alias := range command.GetAliases() {
			c.commands[alias] = command
		}
	}
	c.mu.Unlock()
}

func (c *CommandsManagerImpl) GetCommand(NameOrAlias string) (BaseCommand, bool) {
	c.mu.RLock()
	command, ok := c.commands[NameOrAlias]
	c.mu.RUnlock()
	return command, ok
}
