package interfaces

import (
	"github.com/marouane-souiri/vocalize/internal/domain"
)

type DiscordCacheManager interface {
	/** Guilds Cache */
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

	/** Channels Cache */
	SetChannel(channel *domain.Channel)
	DelChannel(ID string)
	// WARNING:
	// Do not modify the returned *Channel.
	// This object is shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy it before making changes - or we will be fucked.
	GetChannel(ID string) (*domain.Channel, bool)
	ChannelsCount() int

	/** Members Cache */
	SetMember(member *domain.Member)
	DelMember(memberID, guildID string)
	// WARNING:
	// Do not modify the returned *Member.
	// This object is shared between goroutines.
	// Mutating it directly can lead to data races and undefined behavior.
	// Always copy it before making changes - or we will be fucked.
	GetMember(memberID, guildID string) (*domain.Member, bool)
	MembersCount() int
}
