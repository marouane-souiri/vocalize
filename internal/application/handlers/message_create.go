package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/marouane-souiri/vocalize/internal/application/commands/commandsmanager"
	"github.com/marouane-souiri/vocalize/internal/discord/client"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

// https://discord.com/developers/docs/events/gateway-events#message-create
func MessageCreateHandler(c client.Client, commandsManager commandsmanager.CommandsManager) client.HandlerFunc {
	return func(event json.RawMessage) {
		var messageCreate models.MessageCreateEvent

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
