package ws

import (
	"encoding/json"
)

type event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type eventHandler func(event *event, c *websocketClient) error

const (
	eventSendMessage = "send_message"
	eventKraken      = "kraken"
)

type sendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
