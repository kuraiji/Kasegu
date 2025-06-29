package kraken

import (
	"encoding/json"
	"fmt"
	"io"
	"kasegu/external/helpers"
	"reflect"
	"strconv"
	"strings"
)

func (k *kraken) GetAccountBalance() (*map[string]string, error) {
	resp, err := request(&requestParams{
		method:      "POST",
		path:        "/0/private/Balance",
		publicKey:   k.apiKey,
		privateKey:  k.privateKey,
		environment: BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting account balance from endpoint: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading account balance: %w", err)
	}
	var balance struct {
		Error  []string          `json:"error"`
		Result map[string]string `json:"result"`
	}
	if err := json.Unmarshal(data, &balance); err != nil {
		return nil, fmt.Errorf("error parsing the response: %w", err)
	}
	if len(balance.Error) > 0 {
		return nil, fmt.Errorf("error getting account balance: %s", strings.Join(balance.Error, ","))
	}
	return &balance.Result, nil
}

func (k *kraken) GetOHCLData(pair string, interval uint16) (*map[string]any, error) {
	resp, err := request(&requestParams{
		method: "GET",
		path:   "/0/public/OHLC",
		query: map[string]any{
			"pair":     pair,
			"interval": interval,
		},
		environment: BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting OHCL data from endpoint: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading OHCL data: %w", err)
	}
	var ohcl struct {
		Error  []string       `json:"error"`
		Result map[string]any `json:"result"`
	}
	if err := json.Unmarshal(data, &ohcl); err != nil {
		return nil, fmt.Errorf("error parsing OHCL data: %w", err)
	}
	if len(ohcl.Error) > 0 {
		return nil, fmt.Errorf("error getting : %s", strings.Join(ohcl.Error, ","))
	}
	return &ohcl.Result, nil
}

type OHCLData struct {
	Time   float64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Vwap   float64
	Volume float64
	Trades float64
}

func ParseOHCLData(data *map[string]any) (*[]OHCLData, error) {
	var keyName string
	for k := range *data {
		if k != "last" {
			keyName = k
		}
	}
	dataAny := (*data)[keyName]
	dataArr, ok := dataAny.([]any)
	if !ok {
		return nil, fmt.Errorf("error parsing OHCL data")
	}
	ohclData := make([]OHCLData, 0, len(dataArr))
	for i := range dataArr {
		r := reflect.ValueOf(dataArr[i])
		v := make([]float64, r.Len())
		for i := 0; i < r.Len(); i++ {
			var f float64
			var err error
			var ok bool
			f, ok = r.Index(i).Elem().Interface().(float64)
			if !ok {
				f, err = strconv.ParseFloat(r.Index(i).Elem().Interface().(string), 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing OHCL data II: %w", err)
				}
			}
			v[i] = f
		}
		ohclData = append(ohclData, OHCLData{
			Time:   v[0],
			Open:   v[1],
			High:   v[2],
			Low:    v[3],
			Close:  v[4],
			Vwap:   v[5],
			Volume: v[6],
			Trades: v[7],
		})
	}
	return &ohclData, nil
}

func (k *kraken) AddOrder(pair string, volume string, transactionType string) error {
	resp, err := request(&requestParams{
		method:      "POST",
		path:        "/0/private/AddOrder",
		publicKey:   k.apiKey,
		privateKey:  k.privateKey,
		environment: BaseURL,
		body: map[string]any{
			"ordertype": "market",
			"type":      transactionType,
			"pair":      pair,
			"volume":    volume,
		},
	})
	if err != nil {
		return fmt.Errorf("error adding order: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading transaction: %w", err)
	}
	var balance struct {
		Error  []string          `json:"error"`
		Result map[string]string `json:"result"`
	}
	if err := json.Unmarshal(data, &balance); err != nil {
		return fmt.Errorf("error parsing the response: %w", err)
	}
	if len(balance.Error) > 0 {
		return fmt.Errorf("error adding order: %s", strings.Join(balance.Error, ","))
	}
	return nil
}
