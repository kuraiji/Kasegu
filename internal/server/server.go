package server

import (
	"fmt"
	"kasegu/external/helpers"
	"kasegu/internal/kraken"
	tradeBot "kasegu/internal/trade-bot"
	"kasegu/internal/ws"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
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
	tb, err := tradeBot.New()
	fmt.Println(tb)
	if err != nil {
		log.Fatal(err)
	}
	l, err := time.LoadLocation("UTC")
	cr := cron.New(cron.WithLocation(l))
	_, err = cr.AddFunc("1 0 * * *", func() { fmt.Println("Cron job running") })
	if err != nil {
		log.Fatal(err)
	}
	cr.Start()
	wsManager := ws.NewManager(&upgrader)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()
	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.IPExtractor = echo.ExtractIPDirect()
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
