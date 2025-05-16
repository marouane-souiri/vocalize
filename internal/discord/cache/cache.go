package cache

import (
	"sync"

	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

type DiscordCacheManager interface {
	SetGuild(guild *models.Guild)
	DelGuild(guild *models.Guild)
	// WARNING:
	// Do not modify the returned *Guild.
	// This object is shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy it before making changes - or we will be fucked.
	GetGuild(ID string) (*models.Guild, bool)
	// WARNING: Do not modify the values of the returned map (*Guild).
	// The Guild objects are shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy the data before making changes - or we will be fucked.
	GetGuilds() map[string]*models.Guild
	GuildsCount() int
}

type DiscordCacheManagerImpl struct {
	guildsCache   map[string]*models.Guild
	guildsCacheMu sync.RWMutex
}

func NewDiscordCacheManager() DiscordCacheManager {
	return &DiscordCacheManagerImpl{
		guildsCache: make(map[string]*models.Guild, 20),
	}
}

func (c *DiscordCacheManagerImpl) SetGuild(guild *models.Guild) {
	c.guildsCacheMu.Lock()
	c.guildsCache[guild.ID] = guild
	c.guildsCacheMu.Unlock()
}

func (c *DiscordCacheManagerImpl) DelGuild(guild *models.Guild) {
	c.guildsCacheMu.Lock()
	delete(c.guildsCache, guild.ID)
	c.guildsCacheMu.Unlock()
}

func (c *DiscordCacheManagerImpl) GetGuild(ID string) (*models.Guild, bool) {
	c.guildsCacheMu.RLock()
	guild, ok := c.guildsCache[ID]
	c.guildsCacheMu.RUnlock()
	return guild, ok
}

func (c *DiscordCacheManagerImpl) GetGuilds() map[string]*models.Guild {
	c.guildsCacheMu.RLock()
	defer c.guildsCacheMu.RUnlock()

	guilds := make(map[string]*models.Guild, len(c.guildsCache))
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
