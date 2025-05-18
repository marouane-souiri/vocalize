package interfaces

import "github.com/marouane-souiri/vocalize/internal/domain"

type APIRequester interface {
	SendMessage(channelID string, message *domain.SendMessage) error
}
