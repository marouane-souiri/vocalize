package domain

import "time"

type MemberFlags int

const (
	MemberFlags_DID_REJOIN MemberFlags = 1 << iota
	MemberFlags_COMPLETED_ONBOARDING
	MemberFlags_BYPASSES_VERIFICATION
	MemberFlags_STARTED_ONBOARDING
	MemberFlags_IS_GUEST
	MemberFlags_STARTED_HOME_ACTIONS
	MemberFlags_COMPLETED_HOME_ACTIONS
	MemberFlags_AUTOMOD_QUARANTINED_USERNAME
	MemberFlags_DM_SETTINGS_UPSELL_ACKNOWLEDGED
)

type Member struct {
	ID          string      `json:"id"`
	GuildID     string      `json:"guild_id"`
	Nickname    string      `json:"nick"`
	Avatar      string      `json:"avatar"`
	Banner      string      `json:"banner"`
	Roles       []string    `json:"roles"`
	JoinedAt    time.Time   `json:"joined_at"`
	Flags       MemberFlags `json:"flags"`
	Deaf        bool        `json:"deaf"`
	Mute        bool        `json:"mute"`
	Permissions *string     `json:"permissions"`
}
