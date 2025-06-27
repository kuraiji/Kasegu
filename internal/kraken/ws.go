package kraken

import (
	"encoding/json"
	"fmt"
	"kasegu/external/helpers"
	"log"

	"github.com/gorilla/websocket"
)

const (
	PublicWSURL = "wss://ws.kraken.com/v2"
)

type WsClient interface {
	Subscribe(channel string, br BaseRequest) error
	Unsubscribe(channel string) error
	Close() error
	AreThereActiveSubscriptions() bool
	BindResponse() *chan Event
}

type wsClient struct {
	conn     *websocket.Conn
	res      chan Event
	handlers map[string]BaseRequest
	kraken
}

type Event struct {
	Channel   string          `json:"channel"`
	Data      json.RawMessage `json:"data,omitempty"`
	Type      string          `json:"type,omitempty"`
	Timestamp string          `json:"timestamp,omisustempty"`
}

func openConnection(endpoint string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to establish websocket connection to endpoint %s: %w", endpoint, err)
	}
	return c, nil
}

func (wsc *wsClient) readMessages() {
	wsc.res = make(chan Event)
	defer close(wsc.res)
	defer helpers.CheckedClose(wsc)
	for {
		_, msg, err := wsc.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("kraken websocket closed via abnormal closure")
			return
		} else if err != nil {
			log.Printf("kraken websocket read error: %v", err)
			return
		}
		var request Event
		if err := json.Unmarshal(msg, &request); err != nil {
			log.Printf("kraken websocket unmarshal error: %v", err)
			continue
		}
		wsc.res <- request
	}
}

func NewWebSocketClient() (WsClient, error) {
	connection, err := openConnection(PublicWSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to establish websocket connection: %w", err)
	}
	kClient, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to establish kraken client: %w", err)
	}
	wsC := wsClient{
		conn:     connection,
		kraken:   *kClient,
		handlers: make(map[string]BaseRequest),
	}
	go wsC.readMessages()
	return &wsC, nil
}

func (wsc *wsClient) Close() error {
	return helpers.CloseWebsocket(wsc.conn)
}

func (wsc *wsClient) Subscribe(channel string, br BaseRequest) error {
	_, ok := wsc.handlers[channel]
	if ok {
		var err = wsc.Unsubscribe(channel)
		if err != nil {
			log.Printf("kraken websocket unsubscribe error: %v", err)
		}
	}
	if err := br.subscribe(wsc); err != nil {
		return fmt.Errorf("kraken websocket subscription error: %w", err)
	}
	wsc.handlers[channel] = br
	return nil
}

func (wsc *wsClient) Unsubscribe(channel string) error {
	br, ok := wsc.handlers[channel]
	if !ok {
		return fmt.Errorf("kraken websocket subscription error: channel %s not found", channel)
	}
	err := br.unsubscribe(wsc)
	if err != nil {
		return fmt.Errorf("kraken websocket unsubscription error: %w", err)
	}
	delete(wsc.handlers, channel)
	return nil
}

func GetMethodAndChannelFromByteArray(req []byte) (string, string, error) {
	type BasicParam struct {
		Channel string `json:"channel"`
	}
	type BasicRequest struct {
		Method string     `json:"method"`
		Param  BasicParam `json:"params"`
	}
	var bRequest BasicRequest
	if err := json.Unmarshal(req, &bRequest); err != nil {
		return "", "", fmt.Errorf("kraken websocket unmarshal error: %w", err)
	}
	return bRequest.Method, bRequest.Param.Channel, nil
}

func (wsc *wsClient) AreThereActiveSubscriptions() bool {
	return len(wsc.handlers) > 0
}

func (wsc *wsClient) BindResponse() *chan Event {
	return &wsc.res
}

/*
type KrakenEvent struct {
	Channel   string          `json:"channel"`
	Data      json.RawMessage `json:"data,omitempty"`
	Type      string          `json:"type,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

type candlesWSParamsParams struct {
	Channel  string   `json:"channel"`
	Symbol   []string `json:"symbol"`
	Interval uint16   `json:"interval"`
}
type CandlesWSParams struct {
	Method string                `json:"method"`
	Params candlesWSParamsParams `json:"params"`
}

func (wsc *wsClient) CandlesWS(method string, symbol string, interval uint16) error {
	var params = CandlesWSParams{
		Method: method,
		Params: candlesWSParamsParams{
			Channel:  "ohlc",
			Symbol:   []string{symbol},
			Interval: interval,
		},
	}
	err := wsc.conn.WriteJSON(params)
	if err != nil {
		return fmt.Errorf("error sending candles request: %w", err)
	}
	return nil
}*/
