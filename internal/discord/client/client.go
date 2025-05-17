package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/marouane-souiri/vocalize/internal/discord/cache"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
	"github.com/marouane-souiri/vocalize/internal/websocket"
	"github.com/marouane-souiri/vocalize/internal/workerpool"
)

const (
	gatewayURL       = "wss://gateway.discord.gg/?v=10&encoding=json"
	opDispatch       = 0
	opHeartbeat      = 1
	opIdentify       = 2
	opResume         = 6
	opReconnect      = 7
	opInvalidSession = 9
	opHello          = 10
	opHeartbeatACK   = 11
)

type clientHandlerKind int

const (
	clientHandlerNormal       = -1
	clientHandlerOnce         = 1
	clientHandlerConsumedOnce = 0
)

type HandlerFunc func(event json.RawMessage)

type clientHandler struct {
	run  func(event json.RawMessage)
	kind clientHandlerKind
}

type Client interface {
	Start() error
	Stop() error
	On(eventType string, handler HandlerFunc)
	Once(eventType string, handler HandlerFunc)

	SetGuild(guild *models.Guild)
	DelGuild(ID string)

	// WARNING:
	// Do not modify the returned *Guild.
	// This object is shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy it before making changes - or we will be fucked.
	GetGuild(ID string) (*models.Guild, error)
	// WARNING: Do not modify the values of the returned map (*Guild).
	// The Guild objects are shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy the data before making changes - or we will be fucked.
	GetGuilds() map[string]*models.Guild
}

type Payload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
	S  int             `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
}

type clientImpl struct {
	token   string
	url     string
	intents string
	ws      websocket.WSManager
	wp      workerpool.WorkerPool
	cm      cache.DiscordCacheManager

	sessionID        string
	resumeGatewayURL string
	sequence         int

	heartbeatInterval time.Duration
	heartbeatCancel   context.CancelFunc
	lastHeartbeatAck  time.Time

	eventHandlers map[string][]*clientHandler
	mu            sync.RWMutex

	shutdown     chan struct{}
	reconnecting bool
	reconnectMu  sync.Mutex
}

type Intents uint64

const (
	Intents_GUILDS          = 1 << 0
	Intents_GUILD_MESSAGES  = 1 << 9
	Intents_MESSAGE_CONTENT = 1 << 15
)

type CLientOptions struct {
	Ws websocket.WSManager
	Wp workerpool.WorkerPool
	Cm cache.DiscordCacheManager

	Intents uint64
	Token   string
}

func NewClient(options *CLientOptions) (Client, error) {
	if options.Token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	return &clientImpl{
		token:         options.Token,
		url:           gatewayURL,
		intents:       fmt.Sprintf("%d", options.Intents),
		ws:            options.Ws,
		wp:            options.Wp,
		cm:            options.Cm,
		eventHandlers: make(map[string][]*clientHandler),
		shutdown:      make(chan struct{}),
	}, nil
}

func (c *clientImpl) SetGuild(guild *models.Guild) {
	c.cm.SetGuild(guild)
}

func (c *clientImpl) DelGuild(ID string) {
	c.cm.DelGuild(ID)
}

func (c *clientImpl) GetGuild(ID string) (*models.Guild, error) {
	guild, exist := c.cm.GetGuild(ID)
	if !exist {
		// TODO: ask discord api for the guild object
		return nil, errors.New("Guild not found")
	}
	return guild, nil
}

func (c *clientImpl) GetGuilds() map[string]*models.Guild {
	return c.cm.GetGuilds()
}
