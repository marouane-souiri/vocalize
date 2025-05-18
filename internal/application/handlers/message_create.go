package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

// https://discord.com/developers/docs/events/gateway-events#message-create
func MessageCreateHandler(c interfaces.Client, commandsManager interfaces.CommandsManager) domain.ClientHandler {
	return func(event json.RawMessage) {
		var messageCreate domain.MessageCreateEvent

		if err := json.Unmarshal(event, &messageCreate); err != nil {
			log.Printf("[Handlers] Error unmarshaling MESSAGE_CREATE event: %v", err)
			return
		}

		prefix := "."

		if strings.HasPrefix(messageCreate.Content, prefix) {
			cmdName := strings.TrimPrefix(messageCreate.Content, prefix)
			log.Print("command name is: ", cmdName)
			cmd, ok := commandsManager.GetCommand(cmdName)
			if ok {
				cmd.Run(c, struct{}{})
			} else {
				log.Print("Command not found")
			}
		}
	}
}
