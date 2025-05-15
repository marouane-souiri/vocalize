package cache

import (
	"sync"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

type DiscordCacheManager interface {
	Close()

	// guild cache crud funcs
	SetGuild(guild *models.Guild) bool
	SetGuilds(guilds []*models.Guild) bool
	FlushGuilds()
	DelGuild(ID string)
	GetGuild(ID string) (*models.Guild, bool)
	GetGuilds() []*models.Guild
}

type DiscordCacheManagerImpl struct {
	guildsCache *ristretto.Cache[string, *models.Guild]
	guildKeys   sync.Map
}

func NewDiscordCacheManager() (DiscordCacheManager, error) {
	guildsCache, err := ristretto.NewCache(&ristretto.Config[string, *models.Guild]{
		NumCounters: 1e7,
		MaxCost:     1 << 28,
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, err
	}

	return &DiscordCacheManagerImpl{
		guildsCache: guildsCache,
	}, nil
}

func (c *DiscordCacheManagerImpl) Close() {
	c.guildsCache.Close()
}

func (c *DiscordCacheManagerImpl) SetGuild(guild *models.Guild) bool {
	if c.guildsCache.Set(guild.ID, guild, 1) {
		c.guildKeys.Store(guild.ID, struct{}{})
		return true
	}
	return false
}

func (c *DiscordCacheManagerImpl) SetGuilds(guilds []*models.Guild) bool {
	batchSize := 60
	batchCount := 0

	for _, guild := range guilds {
		if !c.SetGuild(guild) {
			return false
		}
		batchCount++

		if batchCount >= batchSize {
			c.FlushGuilds()
			batchCount = 0
		}
	}

	if batchCount > 0 {
		c.FlushGuilds()
	}

	return true
}

func (c *DiscordCacheManagerImpl) FlushGuilds() {
	c.guildsCache.Wait()
}

func (c *DiscordCacheManagerImpl) DelGuild(ID string) {
	c.guildsCache.Del(ID)
	c.guildKeys.Delete(ID)
}

func (c *DiscordCacheManagerImpl) GetGuild(ID string) (*models.Guild, bool) {
	return c.guildsCache.Get(ID)
}

func (c *DiscordCacheManagerImpl) GetGuilds() []*models.Guild {
	var guilds []*models.Guild
	c.guildKeys.Range(func(key, _ any) bool {
		id := key.(string)
		if val, found := c.guildsCache.Get(id); found && val != nil {
			guilds = append(guilds, val)
		}
		return true
	})
	return guilds
}
