package discordcache

import (
	"sync"

	"github.com/marouane-souiri/vocalize/internal/domain"
	"github.com/marouane-souiri/vocalize/internal/interfaces"
)

type DiscordCacheManagerImpl struct {
	guildsCache   map[string]*domain.Guild
	guildsCacheMu sync.RWMutex

	channelsCache   map[string]*domain.Channel
	channelsCacheMu sync.RWMutex
}

func NewDiscordCacheManager() interfaces.DiscordCacheManager {
	return &DiscordCacheManagerImpl{
		guildsCache:   make(map[string]*domain.Guild, 20),
		channelsCache: make(map[string]*domain.Channel, 50),
	}
}

func (c *DiscordCacheManagerImpl) SetGuild(guild *domain.Guild) {
	c.guildsCacheMu.Lock()
	c.guildsCache[guild.ID] = guild
	c.guildsCacheMu.Unlock()
}

func (c *DiscordCacheManagerImpl) DelGuild(ID string) {
	c.guildsCacheMu.Lock()
	delete(c.guildsCache, ID)
	c.guildsCacheMu.Unlock()
}

func (c *DiscordCacheManagerImpl) GetGuild(ID string) (*domain.Guild, bool) {
	c.guildsCacheMu.RLock()
	guild, ok := c.guildsCache[ID]
	c.guildsCacheMu.RUnlock()
	return guild, ok
}

func (c *DiscordCacheManagerImpl) GetGuilds() map[string]*domain.Guild {
	c.guildsCacheMu.RLock()
	defer c.guildsCacheMu.RUnlock()

	guilds := make(map[string]*domain.Guild, len(c.guildsCache))
	for id, guild := range c.guildsCache {
		if !guild.Unavailable {
			guilds[id] = guild
		}
	}
	return guilds
}

func (c *DiscordCacheManagerImpl) GuildsCount() int {
	return len(c.guildsCache)
}

func (c *DiscordCacheManagerImpl) SetChannel(channel *domain.Channel) {
	c.channelsCacheMu.Lock()
	c.channelsCache[channel.ID] = channel
	c.channelsCacheMu.Unlock()
}

func (c *DiscordCacheManagerImpl) DelChannel(ID string) {
	c.channelsCacheMu.Lock()
	delete(c.channelsCache, ID)
	c.channelsCacheMu.Unlock()
}

func (c *DiscordCacheManagerImpl) GetChannel(ID string) (*domain.Channel, bool) {
	c.channelsCacheMu.RLock()
	channel, ok := c.channelsCache[ID]
	c.channelsCacheMu.RUnlock()
	return channel, ok
}

func (c *DiscordCacheManagerImpl) ChannelsCount() int {
	return len(c.channelsCache)
}
