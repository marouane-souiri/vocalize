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

func (c *clientImpl) SetChannel(channel *domain.Channel) {
	c.cm.SetChannel(channel)
}

func (c *clientImpl) DelChannel(ID string) {
	c.cm.DelChannel(ID)
}

func (c *clientImpl) GetChannel(ID string) (*domain.Channel, error) {
	channel, exist := c.cm.GetChannel(ID)
	if !exist {
		// TODO: ask discord api for the Channel object
		return nil, errors.New("Channel not found")
	}
	return channel, nil
}

func (c *clientImpl) SetMember(member *domain.Member) {
	c.cm.SetMember(member)
}

func (c *clientImpl) DelMember(memberID, guildID string) {
	c.cm.DelMember(memberID, guildID)
}

func (c *clientImpl) GetMember(memberID, guildID string) (*domain.Member, error) {
	member, exist := c.cm.GetMember(memberID, guildID)
	if !exist {
		// TODO: ask discord api for the Member object
		return nil, errors.New("Member not found")
	}
	return member, nil
}

func (c *clientImpl) SendMessage(channelID string, message *domain.SendMessage) error {
	c.ar.SendMessage(channelID, message)
	return nil
}
