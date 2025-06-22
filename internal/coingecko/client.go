package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"kasegu/external/helpers"
	"net/http"
)

const envName = "COINGECKO_API_KEY"
const headerName = "x-cg-demo-api-key"
const coingeckoURL = "https://api.coingecko.com/api/v3"

const pingEndpoint = "/ping"

const coinsListEndpoint = "/coins/list"
const coinsListSerializedDataFilename = "coinsList"

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

func CreateClient() (*Coingecko, error) {
	client := &http.Client{}
	envMap, err := helpers.LoadEnv([]string{envName})
	if err != nil {
		return nil, fmt.Errorf("failed retriving api key for coin gecko: %w", err)
	}
	return &Coingecko{apiKey: (*envMap)[envName], client: client}, nil
}
