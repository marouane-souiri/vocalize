package domain

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

type EmbedAuthor struct {
	Name string `json:"name"`
	Url  string `json:"url,omitempty"`
}

type EmbedFooter struct {
	Text string `json:"text"`
	Url  string `json:"url,omitempty"`
}

type EmbedImage struct {
	Url    string `json:"url"`
	Height int    `json:"height,omitempty"`
	Width  int    `json:"widthi,omitempty"`
}

type EmbedThumbnail struct {
	Url    string `json:"url"`
	Height int    `json:"height,omitempty"`
	Width  int    `json:"widthi,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Embed struct {
	Author      EmbedAuthor  `json:"author,omitempty"`
	Title       string       `json:"title,omitempty"`
	Url         string       `json:"url,omitempty"`
	Description string       `json:"description,omitempty"`
	Timestamp   time.Time    `json:"timestamp,omitempty"`
	Color       int          `json:"color,omitempty"`
	Footer      EmbedFooter  `json:"footer,omitempty"`
	Image       EmbedImage   `json:"image,omitempty"`
	Thumbnail   EmbedImage   `json:"thumbnail,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
}

type SendMessage struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}
