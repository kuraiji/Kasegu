package kraken

import (
	"encoding/json"
	"fmt"
	"log"
)

var (
	RequestHandlerMap = map[string]RequestHandler{
		"ohlc": CandleRequest,
	}
)

type BaseRequest interface {
	send(wsc *wsClient) error
	subscribe(wsc *wsClient) error
	unsubscribe(wsc *wsClient) error
}

type RequestHandler func(req json.RawMessage) (BaseRequest, error)

func CandleRequest(req json.RawMessage) (BaseRequest, error) {
	var bReq CandlesRequest
	err := json.Unmarshal(req, &bReq)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return &bReq, nil
}

type baseRequest struct {
	Method string `json:"method"`
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

func (cr *CandlesRequest) subscribe(wsc *wsClient) error {
	cr.Method = "subscribe"
	return cr.send(wsc)
}

func (cr *CandlesRequest) unsubscribe(wsc *wsClient) error {
	cr.Method = "unsubscribe"
	return cr.send(wsc)
}

func (cr *CandlesRequest) send(wsc *wsClient) error {
	log.Println("Called Candles method")
	log.Println(cr)
	err := wsc.conn.WriteJSON(cr)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	return nil
}
