package ws

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type websocketClient struct {
	conn    *websocket.Conn
	manager *websocketManager
}

func newClient(conn *websocket.Conn, wsManager *websocketManager) *websocketClient {
	return &websocketClient{conn: conn, manager: wsManager}
}

func (c *websocketClient) cleanup() {
	c.manager.removeClient(c)
}

func (c *websocketClient) readMessage() {
	defer c.cleanup()
	for {
		_, msg, err := c.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("websocket closed via abnormal closure")
			return
		} else if err != nil {
			log.Printf("websocket read error: %v", err)
			return
		}
		fmt.Println(string(msg))
	}
}

func (c *websocketClient) writeMessage() error {
	return nil
}
