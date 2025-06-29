package data

import (
	"fmt"
	"kasegu/external/helpers"
)

const (
	fileName          = "TBData"
	coingeckoName     = "COINGECKO_API_KEY"
	krakenApiName     = "KRAKEN_API_KEY"
	krakenPrivateName = "KRAKEN_PRIVATE_KEY"
)

var (
	envNames = []string{coingeckoName, krakenApiName, krakenPrivateName}
)

type Data struct {
	CoinGeckoApiKey  string
	KrakenApiKey     string
	KrakenPrivateKey string
	EnableBot        bool
	BaseCurrency     string
	TradingCoin      string
}

func LoadData() (*Data, error) {
	data, err := helpers.UnserializeData[Data](fileName)
	if err == nil {
		fmt.Println("Data loaded from file successfully")
		return data, nil
	}
	envMap, err := helpers.LoadEnv(envNames)
	if err != nil {
		return nil, fmt.Errorf("error loading env variables: %v", err)
	}
	return &Data{
		CoinGeckoApiKey:  (*envMap)[coingeckoName],
		KrakenApiKey:     (*envMap)[krakenApiName],
		KrakenPrivateKey: (*envMap)[krakenPrivateName],
		EnableBot:        true,
		BaseCurrency:     "USD",
		TradingCoin:      "PENGU",
	}, nil
}

func SaveData(data *Data) error {
	err := helpers.SerializeData[Data](data, fileName)
	return err
}
