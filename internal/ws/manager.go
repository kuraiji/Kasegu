package ws

import (
	"fmt"
	"kasegu/external/helpers"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebsocketManager interface{}

type websocketManager struct {
	clients  map[*websocketClient]bool
	upgrader *websocket.Upgrader
	sync.Mutex
}

func NewManager(upgrader *websocket.Upgrader) WebsocketManager {
	return &websocketManager{
		clients:  make(map[*websocketClient]bool),
		upgrader: upgrader,
	}
}

func (wm *websocketManager) ServeWebsocket(c echo.Context) error {
	fmt.Println("websocket request")
	ws, err := wm.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	wc := newClient(ws, wm)
	wm.addClient(wc)
	return nil
}

func (wm *websocketManager) addClient(wsClient *websocketClient) {
	wm.Lock()
	defer wm.Unlock()
	wm.clients[wsClient] = true
}

func (wm *websocketManager) removeClient(wsClient *websocketClient) {
	wm.Lock()
	defer wm.Unlock()
	if _, ok := wm.clients[wsClient]; ok {
		helpers.CheckedWSClose(wsClient.conn)
		delete(wm.clients, wsClient)
	}
}
