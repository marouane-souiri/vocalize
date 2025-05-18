package domain

import "encoding/json"

type ClientHandler func(event json.RawMessage)
