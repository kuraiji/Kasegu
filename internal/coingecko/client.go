package coingecko

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kasegu/external/helpers"
	"net/http"
	"strings"
)

const envName = "COINGECKO_API_KEY"
const headerName = "x-cg-demo-api-key"
const coingeckoURL = "https://api.coingecko.com/api/v3"

const pingEndpoint = "/ping"

const coinsListEndpoint = "/coins/list"
const coinsListSerializedDataFilename = "coinsList"

const coinHistoricalChartEndpoint = "/coins/%s/market_chart/range"
const coinHistoricalChartSerializedDataFilename = "coinHistoricalChart"

type Coingecko struct {
	client *http.Client
	apiKey string
}

type CoinData struct {
	Id       string                 `json:"id"`
	Symbol   string                 `json:"symbol"`
	Name     string                 `json:"name"`
	Platform map[string]interface{} `json:"platform"`
}

func (c *Coingecko) Ping() (string, error) {
	url := fmt.Sprintf("%s%s", coingeckoURL, pingEndpoint)
	resp, err := helpers.GetWithHeaders(c.client, url, http.Header{headerName: {c.apiKey}})
	if err != nil {
		return "", fmt.Errorf("failed pinging coingecko API: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading the response from the endpoint: %w", err)
	}
	return string(body), nil
}

func (c *Coingecko) CoinsList() (*[]CoinData, error) {
	ok := helpers.IsThereSerializedData(coinsListSerializedDataFilename)
	if ok {
		fmt.Println("Attempting to Fetch Coins Data from Local Disk...")
		data, err := helpers.UnserializeData[[]CoinData](coinsListSerializedDataFilename)
		if err != nil {
			return nil, fmt.Errorf("failed unserializing the response from local file: %w", err)
		}
		return data, nil
	}
	fmt.Println("Attempting to connect to CG API to Fetch Coins...")
	url := fmt.Sprintf("%s%s", coingeckoURL, coinsListEndpoint)
	resp, err := helpers.GetWithHeaders(c.client, url, http.Header{headerName: {c.apiKey}})
	if err != nil {
		return nil, fmt.Errorf("failed connecting to coingecko API: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading the response from the endpoint: %w", err)
	}
	var coins []CoinData
	if err := json.Unmarshal(body, &coins); err != nil {
		return nil, fmt.Errorf("failed parsing the response from the endpoint: %w", err)
	}
	err = helpers.SerializeData(&coins, coinsListSerializedDataFilename)
	if err != nil {
		return nil, fmt.Errorf("failed serializing the data: %w", err)
	}
	return &coins, nil
}

type GetCoinDataParams struct {
	Symbol string
	Id     string
	Name   string
	Coins  *[]CoinData
}

func GetCoinData(params *GetCoinDataParams) (*CoinData, error) {
	if params.Coins == nil {
		return nil, errors.New("no coin list provided")
	}
	if params.Id == "" && params.Name == "" && params.Symbol == "" {
		return nil, errors.New("no filters provided")
	}
	for _, coin := range *params.Coins {
		if params.Symbol != "" && strings.ToLower(coin.Symbol) != strings.ToLower(params.Symbol) {
			continue
		}
		if params.Id != "" && strings.ToLower(coin.Id) != strings.ToLower(params.Id) {
			continue
		}
		if params.Name != "" && strings.ToLower(coin.Name) != strings.ToLower(params.Name) {
			continue
		}
		return &coin, nil
	}
	return nil, errors.New("couldn't find coin in list")
}

type GetCoinHistoricalChartParams struct {
	VsCurrency string
	From       int64
	To         int64
	Interval   string
	Precision  string
	Coin       *CoinData
}

func (c *Coingecko) CoinHistoricalChart(params *GetCoinHistoricalChartParams) (string, error) {
	if params.Coin == nil || params.From == 0 || params.To == 0 {
		return "", errors.New("required parameters not provided")
	}
	if ok := helpers.IsThereSerializedData(coinHistoricalChartSerializedDataFilename); ok {
		fmt.Println("Attempting to Fetch Historical Chart Data from Local Disk...")
		//data, err := helpers.UnserializeData()
	}
	fmt.Println("Attempting to connect to CG API to Fetch Historical Chart Data...")
	url := fmt.Sprintf("%s%s", coingeckoURL, fmt.Sprintf(coinHistoricalChartEndpoint, params.Coin.Id))
	qParamMap := map[string]string{}
	qParamMap["from"] = fmt.Sprintf("%d", params.From)
	qParamMap["to"] = fmt.Sprintf("%d", params.To)
	if params.VsCurrency == "" {
		qParamMap["vs_currency"] = "usd"
	} else {
		qParamMap["vs_currency"] = params.VsCurrency
	}
	if params.Interval != "" {
		qParamMap["interval"] = params.Interval
	}
	if params.Precision != "" {
		qParamMap["precision"] = params.Precision
	}
	url = helpers.AppendQueryParameters(url, &qParamMap)
	fmt.Println(url)
	resp, err := helpers.GetWithHeaders(c.client, url, http.Header{headerName: {c.apiKey}})
	if err != nil {
		return "", fmt.Errorf("failed connecting to congecko API: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading the response from the endpoint: %w", err)
	}
	return fmt.Sprintf("%s", string(body)), nil
}

func CreateClient() (*Coingecko, error) {
	client := &http.Client{}
	envMap, err := helpers.LoadEnv([]string{envName})
	if err != nil {
		return nil, fmt.Errorf("failed retriving api key for coin gecko: %w", err)
	}
	return &Coingecko{apiKey: (*envMap)[envName], client: client}, nil
}
