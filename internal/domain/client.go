package domain

import "encoding/json"

type ClientHandler func(event json.RawMessage)

type Intents uint64

const (
	Intents_GUILDS          = 1 << 0
	Intents_GUILD_MESSAGES  = 1 << 9
	Intents_MESSAGE_CONTENT = 1 << 15
)
