package ws

import (
	"fmt"
	"kasegu/external/helpers"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebsocketManager interface {
	ServeWebsocket(c echo.Context) error
}

type websocketManager struct {
	clients  map[*websocketClient]bool
	upgrader *websocket.Upgrader
	sync.Mutex
	handlers map[string]eventHandler
}

func NewManager(upgrader *websocket.Upgrader) WebsocketManager {
	m := &websocketManager{
		clients:  make(map[*websocketClient]bool),
		upgrader: upgrader,
		handlers: make(map[string]eventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (wm *websocketManager) ServeWebsocket(c echo.Context) error {
	fmt.Println("websocket request")
	ws, err := wm.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	wc := newClient(ws, wm)
	wm.addClient(wc)
	go wc.readMessages()
	go wc.writeMessages()
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

func (wm *websocketManager) setupEventHandlers() {
	wm.handlers[eventSendMessage] = sendMessage
}

func sendMessage(event *event, c *websocketClient) error {
	fmt.Println(*event)
	return nil
}

func (wm *websocketManager) routeEvent(e *event, c *websocketClient) error {
	if handler, ok := wm.handlers[e.Type]; ok {
		if err := handler(e, c); err != nil {
			return fmt.Errorf("handle event %s error: %v", e.Type, err)
		}
		return nil
	} else {
		return fmt.Errorf("unknown event type %s", e.Type)
	}
}
