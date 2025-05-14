package cache

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

type DiscordCacheManager interface {
	GetGuild(ID string) (*models.Guild, bool)
	SetGuild(ID string, guild *models.Guild) bool
}

type DiscordCacheManagerImpl struct {
	guildsCache *ristretto.Cache[string, *models.Guild]
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

func (c *DiscordCacheManagerImpl) GetGuild(ID string) (*models.Guild, bool) {
	return c.guildsCache.Get(ID)
}

func (c *DiscordCacheManagerImpl) SetGuild(ID string, guild *models.Guild) bool {
	return c.guildsCache.Set(ID, guild, 1)
}
