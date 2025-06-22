package ibkr

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"kasegu/external/helpers"
	"net/http"
)

const ibkrUrl = "https://localhost:5000/v1/api"
const authStatusEndpoint = "/iserver/auth/status"

type Ibkr struct {
	client *http.Client
}

func (i *Ibkr) IsAuthenticated() (bool, error) {
	url := fmt.Sprintf("%s%s", ibkrUrl, authStatusEndpoint)
	resp, err := i.client.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed connecting to endpoint: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed reading the response from the endpoint: %w", err)
	}
	var result struct {
		Authenticated bool `json:"authenticated"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed parsing the response from the endpoint: %w", err)
	}

	return result.Authenticated, nil
}

func CreateClient() *Ibkr {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	return &Ibkr{client: client}
}
