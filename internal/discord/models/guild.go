package models

type UnavailableGuild struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable"`
}

type DefaultMessageNotificationLevel int

const (
	ALL_MESSAGES DefaultMessageNotificationLevel = iota
	ONLY_MENTIONS
)

type VerificationLevel int

const (
	VerificationLevel_NONE VerificationLevel = iota
	VerificationLevel_LOW
	VerificationLevel_MEDIUM
	VerificationLevel_HIGH
	VerificationLevel_VERY_HIGH
)

type ExplicitContentFilterLevel int

const (
	ExplicitContentFilterLevel_DISABLED ExplicitContentFilterLevel = iota
	ExplicitContentFilterLevel_MEMBERS_WITHOUT_ROLES
	ExplicitContentFilterLevel_ALL_MEMBERS
)

type MFALevel int

const (
	MFALevel_NONE MFALevel = iota
	MFALevel_ELEVATED
)

type PremiumTier int

const (
	PremiumTier_NONE PremiumTier = iota
	PremiumTier_TIER_1
	PremiumTier_TIER_2
	PremiumTier_TIER_3
)

type Guild struct {
	ID                          string                          `json:"id"`
	Unavailable                 bool                            `json:"unavailable"`
	Name                        string                          `json:"name"`
	Icon                        *string                         `json:"icon"`
	OwnerID                     string                          `json:"owner_id"`
	AfkChannelID                *string                         `json:"afk_channel_id"`
	AfkTimeout                  int                             `json:"afk_timeout"`
	VerificationLevel           VerificationLevel               `json:"verification_level"`
	DefaultMessageNotifications DefaultMessageNotificationLevel `json:"default_message_notifications"`
	ExplicitContentFilter       ExplicitContentFilterLevel      `json:"explicit_content_filter"`
	MfaLevel                    MFALevel                        `json:"mfa_level"`
	SystemChannelID             *string                         `json:"system_channel_id"`
	SystemChannelFlags          int                             `json:"system_channel_flags"`
	RulesChannelID              *string                         `json:"rules_channel_id"`
	MaxPresences                *int                            `json:"max_presences"`
	MaxMembers                  *int                            `json:"max_members"`
	VanityURLCode               *string                         `json:"vanity_url_code"`
	Description                 *string                         `json:"description"`
	Banner                      *string                         `json:"banner"`
	PremiumTier                 int                             `json:"premium_tier"`
	PremiumSubscriptionCount    *int                            `json:"premium_subscription_count"`
	PreferredLocale             string                          `json:"preferred_locale"`
	PublicUpdatesChannelID      *string                         `json:"public_updates_channel_id"`
	MaxVideoChannelUsers        *int                            `json:"max_video_channel_users"`
	MaxStageVideoChannelUsers   *int                            `json:"max_stage_video_channel_users"`
	NsfwLevel                   int                             `json:"nsfw_level"`
	PremiumProgressBarEnabled   bool                            `json:"premium_progress_bar_enabled"`
	SafetyAlertsChannelID       *string                         `json:"safety_alerts_channel_id"`
}
