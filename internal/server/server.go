package server

import (
	"kasegu/external/helpers"
	"kasegu/internal/kraken"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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
	e := echo.New()
	if (*envs)["ENV"] == "development" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"http://localhost:3000"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
	}
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
