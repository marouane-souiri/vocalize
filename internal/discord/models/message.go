package models

import "time"

type MessageType int

const (
	MessageType_DEFAULT MessageType = iota
)

type Message struct {
	ID              string      `json:"id"`
	ChannelID       string      `json:"channel_id"`
	Author          User        `json:"author"`
	Content         string      `json:"content"`
	Timestamp       time.Time   `json:"timestamp"`
	EditedTimestamp *time.Time  `json:"edited_timestamp"`
	Type            MessageType `json:"type"`
}
