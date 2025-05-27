package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
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

type clientHandler struct {
	run  domain.ClientHandler
	kind clientHandlerKind
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
	ws      interfaces.WSManager
	wp      interfaces.WorkerPool
	cm      interfaces.DiscordCacheManager
	ar      interfaces.APIRequester

	authenticated bool
	authMu        sync.Mutex

	sessionID         string
	resumeGatewayURL  string
	sequence          int
	heartbeatInterval time.Duration
	heartbeatCancel   context.CancelFunc
	lastHeartbeatAck  time.Time
	eventHandlers     map[string][]*clientHandler
	mu                sync.RWMutex
	shutdown          chan struct{}
	reconnecting      bool
	reconnectMu       sync.Mutex
}

type Intents uint64

const (
	Intents_GUILDS          = 1 << 0
	Intents_GUILD_MESSAGES  = 1 << 9
	Intents_MESSAGE_CONTENT = 1 << 15
)

type CLientOptions struct {
	Ws interfaces.WSManager
	Wp interfaces.WorkerPool
	Cm interfaces.DiscordCacheManager
	Ar interfaces.APIRequester

	Intents uint64
	Token   string
}

func NewClient(options *CLientOptions) (interfaces.Client, error) {
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
		ar:            options.Ar,
		eventHandlers: make(map[string][]*clientHandler),
		shutdown:      make(chan struct{}),
		authenticated: false,
	}, nil
}

func (c *clientImpl) SetGuild(guild *domain.Guild) {
	c.cm.SetGuild(guild)
}

func (c *clientImpl) DelGuild(ID string) {
	c.cm.DelGuild(ID)
}

func (c *clientImpl) GetGuild(ID string) (*domain.Guild, error) {
	guild, exist := c.cm.GetGuild(ID)
	if !exist {
		// TODO: ask discord api for the guild object
		return nil, errors.New("Guild not found")
	}
	return guild, nil
}

func (c *clientImpl) GetGuilds() map[string]*domain.Guild {
	return c.cm.GetGuilds()
}

func (c *clientImpl) SendMessage(channelID string, message *domain.SendMessage) error {
	c.ar.SendMessage(channelID, message)
	return nil
}
