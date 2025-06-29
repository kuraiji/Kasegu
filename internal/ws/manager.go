package ws

import (
	"context"
	"fmt"
	"kasegu/external/helpers"
	"kasegu/internal/data"
	"kasegu/internal/kraken"
	"log"
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
	ips      map[string]*websocketClient
	upgrader *websocket.Upgrader
	sync.Mutex
	evHandlers map[string]eventHandler
	tbd        *data.Data
}

func NewManager(upgrader *websocket.Upgrader, d *data.Data) WebsocketManager {
	m := &websocketManager{
		clients:    make(map[*websocketClient]bool),
		ips:        make(map[string]*websocketClient),
		upgrader:   upgrader,
		evHandlers: make(map[string]eventHandler),
		tbd:        d,
	}
	m.setupEventHandlers()
	return m
}

func (wm *websocketManager) ServeWebsocket(c echo.Context) error {
	fmt.Println("websocket request")
	ip := c.RealIP()
	oldWsc, ok := wm.doesClientAlreadyExist(ip)
	if ok {
		log.Println("old client being found and deleting it")
		oldWsc.cleanup()
	}
	ws, err := wm.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	wc := newClient(ws, wm)
	wm.addClient(wc, ip)
	go wc.readMessages()
	go wc.writeMessages()
	return nil
}

func (wm *websocketManager) addClient(wsClient *websocketClient, ip string) {
	wm.Lock()
	defer wm.Unlock()
	wm.clients[wsClient] = true
	wm.ips[ip] = wsClient
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
	wm.evHandlers[eventSendMessage] = sendMessage
	wm.evHandlers[eventKraken] = sendKraken
}

func sendMessage(event *event, c *websocketClient) error {
	fmt.Println(*event)
	return nil
}

func sendKraken(event *event, c *websocketClient) error {
	method, channel, err := kraken.GetMethodAndChannelFromByteArray(event.Payload)
	if err != nil {
		return err
	}
	br, ok := kraken.RequestHandlerMap[channel]
	if !ok {
		return fmt.Errorf("channel %s not supported", channel)
	}
	msg, err := br(event.Payload)
	if err != nil {
		return err
	}
	if method == "subscribe" {
		if c.kClient == nil {
			err := c.createKrakenClient(c.manager.tbd.KrakenApiKey, c.manager.tbd.KrakenApiKey)
			if err != nil {
				return fmt.Errorf("create kraken client error: %v", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			go c.deliverKrakenEvents(ctx)
			c.cancel = &cancel
		}
		err := (*c.kClient).Subscribe(channel, msg)
		if err != nil {
			return fmt.Errorf("subscribe error: %v", err)
		}
	} else if method == "unsubscribe" {
		if c.kClient == nil {
			log.Printf("Kraken client doesn't exist")
			return nil
		}
		err := (*c.kClient).Unsubscribe(channel)
		if err != nil {
			return fmt.Errorf("unsubscribe error: %v", err)
		}
		if !(*c.kClient).AreThereActiveSubscriptions() {
			c.destroyKrakenClient()
		}
	} else {
		return fmt.Errorf("method %s not supported", method)
	}
	return nil
}

func (wm *websocketManager) routeEvent(e *event, c *websocketClient) error {
	if handler, ok := wm.evHandlers[e.Type]; ok {
		if err := handler(e, c); err != nil {
			return fmt.Errorf("handle event %s error: %v", e.Type, err)
		}
		return nil
	} else {
		return fmt.Errorf("unknown event type %s", e.Type)
	}
}

func (wm *websocketManager) doesClientAlreadyExist(ip string) (*websocketClient, bool) {
	if wsc, ok := wm.ips[ip]; ok {
		return wsc, true
	}
	return nil, false
}
