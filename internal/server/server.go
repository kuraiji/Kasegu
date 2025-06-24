package server

import (
	"fmt"
	"kasegu/external/helpers"
	"kasegu/internal/kraken"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const devUrl = "http://localhost:3000"

func Loop() {
	envs, err := helpers.LoadEnv([]string{"ENV"})
	if err != nil {
		log.Fatal(err)
	}
	kClient, err := kraken.New()
	if err != nil {
		log.Fatal(err)
	}
	defer helpers.CheckedClose(kClient)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		helpers.CheckedClose(kClient)
		os.Exit(1)
	}()
	upgrader := websocket.Upgrader{}
	e := echo.New()
	if (*envs)["ENV"] == "development" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{devUrl},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
		upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == devUrl
		}
	}
	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.GET("/ws", func(c echo.Context) error { return websocketRequest(c, upgrader) })
	e.GET("/api/chart", func(c echo.Context) error { return getChart(c, &kClient) })
	e.Logger.Fatal(e.Start(":1323"))
}

func websocketRequest(c echo.Context, upgrader websocket.Upgrader) error {
	fmt.Println("websocket request")
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer helpers.CheckedWSClose(ws)
	for {
		err := ws.WriteMessage(websocket.TextMessage, []byte("hello world"))
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			fmt.Println("websocket closed")
			return nil
		} else if err != nil {
			c.Logger().Error(err)
			return nil
		}
		_, msg, err := ws.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			fmt.Println("websocket closed")
			return nil
		} else if err != nil {
			c.Logger().Error(err)
			return nil
		}
		fmt.Println(string(msg))
	}
}

func getChart(c echo.Context, k *kraken.Kraken) error {
	pair := c.QueryParam("pair")
	if pair == "" {
		return c.String(http.StatusBadRequest, "pair is required")
	}
	interval := c.QueryParam("interval")
	if interval == "" {
		return c.String(http.StatusBadRequest, "interval is required")
	}
	i, err := strconv.ParseInt(interval, 10, 16)
	if err != nil {
		return c.String(http.StatusBadRequest, "interval needs to be an integer")
	}
	data, err := (*k).GetOHCLData(pair, uint16(i))
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed getting data")
	}
	return c.JSON(http.StatusOK, data)
}
