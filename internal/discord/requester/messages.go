package requester

import (
	"fmt"

	"github.com/marouane-souiri/vocalize/internal/discord/models"
)

func (api *APIRequesterImpl) SendMessage(channelID string, message *models.SendMessage) error {
	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)
	return api.BaseReq("POST", endpoint, message, nil)
}
