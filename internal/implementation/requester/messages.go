package requester

import (
	"fmt"

	"github.com/marouane-souiri/vocalize/internal/domain"
)

func (api *APIRequesterImpl) SendMessage(channelID string, message *domain.SendMessage) error {
	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)
	return api.BaseReq("POST", endpoint, message, nil)
}
