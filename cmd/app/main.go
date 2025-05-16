package main

import (
	"log"

	"github.com/marouane-souiri/vocalize/internal/application/commands/commands"
	"github.com/marouane-souiri/vocalize/internal/application/commands/commandsmanager"
	"github.com/marouane-souiri/vocalize/internal/application/handlers"
	"github.com/marouane-souiri/vocalize/internal/config"
	"github.com/marouane-souiri/vocalize/internal/discord/cache"
	"github.com/marouane-souiri/vocalize/internal/discord/client"
	"github.com/marouane-souiri/vocalize/internal/discord/websocket"
	"github.com/marouane-souiri/vocalize/internal/workerpool"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	websocketManager := websocket.NewWSManager()
	defer websocketManager.Close()

	workerpoolManager := workerpool.NewWorkerPool(10, 5)
	defer workerpoolManager.Shutdown(nil)

	discordCacheManager := cache.NewDiscordCacheManager()

	client, err := client.NewClient(&client.CLientOptions{
		Token:   config.Conf.Discord.Token,
		Intents: client.Intents_GUILDS | client.Intents_GUILD_MESSAGES | client.Intents_MESSAGE_CONTENT,
		Ws:      websocketManager,
		Wp:      workerpoolManager,
		Cm:      discordCacheManager,
	})
	if err != nil {
		log.Fatalf("Failed to create discord client: %v", err)
	}

	commandsManager := commandsmanager.NewCommandsManager()

	commandsManager.AddCommand(commands.NewPingCommand())

	client.On("GUILD_CREATE", handlers.GuildCreateHandler(client))
	client.On("GUILD_DELETE", handlers.GuildDeleteHandler(client))
	client.On("MESSAGE_CREATE", handlers.MessageCreateHandler(client, commandsManager))

	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start discord client: %v", err)
	}
	defer client.Stop()

	select {}
}
