package kraken

import (
	"encoding/json"
	"fmt"
	"log"
)

type BaseRequest interface {
	send(wsc *wsClient) error
	subscribe(wsc *wsClient) error
	unsubscribe(wsc *wsClient) error
}

type baseRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

func (br *baseRequest) subscribe(wsc *wsClient) error {
	br.Method = "subscribe"
	return br.send(wsc)
}

func (br *baseRequest) unsubscribe(wsc *wsClient) error {
	br.Method = "unsubscribe"
	return br.send(wsc)
}

func (br *baseRequest) send(wsc *wsClient) error {
	log.Println(br.Params)
	err := wsc.conn.WriteJSON(br)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	return nil
}

type CandlesParams struct {
	Channel  string   `json:"channel"`
	Symbol   []string `json:"symbol"`
	Interval uint16   `json:"interval"`
}
type CandlesRequest struct {
	baseRequest
	Params CandlesParams `json:"params"`
}
