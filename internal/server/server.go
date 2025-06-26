package server

import (
	"kasegu/external/helpers"
	"kasegu/internal/kraken"
	"kasegu/internal/ws"
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
	kClient, err := kraken.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	e := echo.New()
	if (*envs)["ENV"] == "development" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{devUrl},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	wsManager := ws.NewManager(&upgrader)
	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.GET("/ws", wsManager.ServeWebsocket)
	e.GET("/api/chart", func(c echo.Context) error { return getChart(c, &kClient) })
	e.Logger.Fatal(e.Start(":1323"))
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
