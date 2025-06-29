package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"kasegu/external/helpers"
	"kasegu/internal/kraken"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type websocketClient struct {
	conn    *websocket.Conn
	manager *websocketManager
	egress  chan event
	kClient *kraken.WsClient
	cancel  *context.CancelFunc
}

func newClient(conn *websocket.Conn, wsManager *websocketManager) *websocketClient {
	return &websocketClient{conn: conn, manager: wsManager, egress: make(chan event)}
}

func (c *websocketClient) cleanup() {
	if c.kClient != nil {
		c.destroyKrakenClient()
	}
	c.manager.removeClient(c)
}

func (c *websocketClient) readMessages() {
	defer c.cleanup()
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("Error setting read deadline:", err)
		return
	}
	c.conn.SetReadLimit(1024)
	c.conn.SetPongHandler(c.pongHandler)
	for {
		_, msg, err := c.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("websocket closed via abnormal closure")
			return
		} else if err != nil {
			log.Printf("websocket read error: %v", err)
			return
		}
		var request event
		if err := json.Unmarshal(msg, &request); err != nil {
			log.Printf("websocket unmarshal error: %v", err)
			continue
		}
		if err := c.manager.routeEvent(&request, c); err != nil {
			log.Printf("websocket routing error: %v", err)
			continue
		}
	}
}

func (c *websocketClient) writeMessages() {
	defer c.cleanup()
	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case msg, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Printf("connection closed: %v\n", err)
				}
				return
			}
			marshal, err := json.Marshal(msg)
			if err != nil {
				log.Printf("json marshal error: %v\n", err)
				continue
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, marshal); err != nil {
				log.Printf("failed to send message: %v\n", err)
			}
			log.Println("message sent")
		case <-ticker.C:
			log.Println("ping interval")
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("failed to send ping: %v\n", err)
				return
			}
		}
	}
}

func (c *websocketClient) createKrakenClient(apiKeyEnv string, privateKeyEnv string) error {
	client, err := kraken.NewWebSocketClient(apiKeyEnv, privateKeyEnv)
	if err != nil {
		return fmt.Errorf("failed to create kraken websocket client: %w", err)
	}
	c.kClient = &client
	return nil
}

func (c *websocketClient) destroyKrakenClient() {
	if c.cancel != nil {
		(*c.cancel)()
		c.cancel = nil
	}
	helpers.CheckedClose(*c.kClient)
	c.kClient = nil
	return
}

func (c *websocketClient) deliverKrakenEvents(ctx context.Context) {
	for {
		select {
		case ev, ok := <-(*(*c.kClient).BindResponse()):
			if !ok {
				log.Println("kraken event delivery channel closed")
				return
			}
			evJson, err := json.Marshal(ev)
			if err != nil {
				log.Printf("json marshal error: %v\n", err)
				continue
			}
			c.egress <- event{
				Type:    eventKraken,
				Payload: evJson,
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *websocketClient) pongHandler(_ string) error {
	log.Println("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
