package server

import (
	"fmt"
	"kasegu/internal/kraken"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

//Market Making Bot
//Macro News
//Specific Stock News
//Futures before open
//RSI and AO
//Candlestick Patterns
//Heikin-Ashi Chart (Moving Average)

func Loop() {
	/*ibkrClient := ibkr.CreateClient()
	isAuthenticated, err := ibkrClient.IsAuthenticated()
	if err != nil || !isAuthenticated {
		log.Fatal(fmt.Errorf("client isn't authenticated: %v", err))
	}*/
	/*cgClient, err := coingecko.CreateClient()
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cgClient.CoinsList()
	if err == nil {
		fmt.Println((*resp)[0].Name)
	} else {
		log.Fatal(err)
	}
	coin, err := coingecko.GetCoinData(&coingecko.GetCoinDataParams{Coins: resp, Name: "bitcoin"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(coin)
	histData, err := cgClient.CoinHistoricalChart(&coingecko.GetCoinHistoricalChartParams{
		Coin: coin,
		To:   time.Now().Unix(),
		From: time.Now().AddDate(-1, 0, 0).Unix(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(histData)*/

	kClient, err := kraken.New()
	if err != nil {
		log.Fatal(err)
	}
	b, err := kClient.GetAccountBalance()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(b)

	e := echo.New()
	e.GET("/", func(c echo.Context) error { return c.JSON(http.StatusOK, "Hello, World!") })
	e.Logger.Fatal(e.Start(":1323"))
}
