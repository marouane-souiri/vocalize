package interfaces

import (
	"github.com/marouane-souiri/vocalize/internal/domain"
)

type DiscordCacheManager interface {
	SetGuild(guild *domain.Guild)
	DelGuild(ID string)
	// WARNING:
	// Do not modify the returned *Guild.
	// This object is shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy it before making changes - or we will be fucked.
	GetGuild(ID string) (*domain.Guild, bool)
	// WARNING: Do not modify the values of the returned map (*Guild).
	// The Guild objects are shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy the data before making changes - or we will be fucked.
	GetGuilds() map[string]*domain.Guild
	GuildsCount() int
}
