package kraken

import (
	"encoding/json"
	"fmt"
	"io"
	"kasegu/external/helpers"
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
