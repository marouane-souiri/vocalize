package main

import (
	"log"

	"github.com/marouane-souiri/vocalize/internal/config"
	"github.com/marouane-souiri/vocalize/internal/domain"

	"github.com/marouane-souiri/vocalize/internal/implementation/client"
	"github.com/marouane-souiri/vocalize/internal/implementation/commandscontext"
	"github.com/marouane-souiri/vocalize/internal/implementation/commandsmanager"
	"github.com/marouane-souiri/vocalize/internal/implementation/discordcache"
	"github.com/marouane-souiri/vocalize/internal/implementation/ratelimiter"
	"github.com/marouane-souiri/vocalize/internal/implementation/requester"
	"github.com/marouane-souiri/vocalize/internal/implementation/websocket"
	"github.com/marouane-souiri/vocalize/internal/implementation/workerpool"

	"github.com/marouane-souiri/vocalize/internal/application/commands"
	"github.com/marouane-souiri/vocalize/internal/application/handlers"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	websocketManager := websocket.NewWSManager()
	defer websocketManager.Close()

	workerpoolManager := workerpool.NewWorkerPool(10, 5)
	defer workerpoolManager.Shutdown(nil)

	discordCacheManager := discordcache.NewDiscordCacheManager()

	rateLimiter := ratelimiter.NewRateLimiter()

	apiRequester := requester.NewAPIRequester(config.Conf.Discord.Token, rateLimiter)

	client, err := client.NewClient(&client.CLientOptions{
		Token:   config.Conf.Discord.Token,
		Intents: domain.Intents_GUILDS | domain.Intents_GUILD_MESSAGES | domain.Intents_MESSAGE_CONTENT,
		Ws:      websocketManager,
		Wp:      workerpoolManager,
		Cm:      discordCacheManager,
		Ar:      apiRequester,
	})
	if err != nil {
		log.Fatalf("Failed to create discord client: %v", err)
	}

	commandsContextMaker := commandscontext.NewCommandsContextMaker()
	commandsManager := commandsmanager.NewCommandsManager()

	commandsManager.AddCommand(commands.NewPingCommand())

	client.On("GUILD_CREATE", handlers.GuildCreateHandler(client))
	client.On("GUILD_UPDATE", handlers.GuildUpdateHandler(client))
	client.On("GUILD_DELETE", handlers.GuildDeleteHandler(client))

	client.On("CHANNEL_CREATE", handlers.ChannelCreateHandler(client))
	client.On("CHANNEL_UPDATE", handlers.ChannelUpdateHandler(client))
	client.On("CHANNEL_DELETE", handlers.ChannelDeleteHandler(client))

	client.On("GUILD_MEMBER_ADD", handlers.MemberAddHandler(client))
	client.On("GUILD_MEMBER_REMOVE", handlers.MemberRemoveHandler(client))
	client.On("GUILD_MEMBER_UPDATE", handlers.MemberUpdateHandler(client))

	client.On("MESSAGE_CREATE", handlers.MessageCreateHandler(client, commandsManager, commandsContextMaker))

	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start discord client: %v", err)
	}
	defer client.Stop()

	select {}
}
