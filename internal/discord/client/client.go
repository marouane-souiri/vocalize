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
	"github.com/marouane-souiri/vocalize/internal/discord/websocket"
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

type Client interface {
	Start() error
	Stop() error
	On(eventType string, handler func(event json.RawMessage))
	Once(eventType string, handler func(event json.RawMessage))

	GetGuild(ID string) (*models.Guild, error)
	GetGuilds() []*models.Guild
}

type Payload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
	S  int             `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
}

type clientHandlerKind int

const (
	clientHandlerNormal       = -1
	clientHandlerOnce         = 1
	clientHandlerConsumedOnce = 0
)

type clientHandler struct {
	run  func(event json.RawMessage)
	kind clientHandlerKind
}

type clientImpl struct {
	token string
	url   string
	ws    websocket.WSManager
	wp    workerpool.WorkerPool
	cm    cache.DiscordCacheManager

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

type CLientOptions struct {
	Ws websocket.WSManager
	Wp workerpool.WorkerPool
	Cm cache.DiscordCacheManager

	Token string
}

func NewClient(options *CLientOptions) (Client, error) {
	if options.Token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	return &clientImpl{
		token:         options.Token,
		url:           gatewayURL,
		ws:            options.Ws,
		wp:            options.Wp,
		cm:            options.Cm,
		eventHandlers: make(map[string][]*clientHandler),
		shutdown:      make(chan struct{}),
	}, nil
}

func (c *clientImpl) GetGuild(ID string) (*models.Guild, error) {
	guild, exist := c.cm.GetGuild(ID)
	if !exist {
		return nil, errors.New("Guild not found")
	}
	return guild, nil
}

func (c *clientImpl) GetGuilds() []*models.Guild {
	return c.cm.GetGuilds()
}
