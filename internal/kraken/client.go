package kraken

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"kasegu/external/helpers"
	"net/http"
	"time"
)

const (
	BaseURL   = "https://api.kraken.com"
	UserAgent = "KaseguKrakenClient/0.1.0"
)

type Kraken interface {
	GetAccountBalance() (*map[string]string, error)
	GetOHCLData(pair string, interval uint16) (*map[string]any, error)
}
type kraken struct {
	apiKey     string
	privateKey string
}

type requestParams struct {
	method      string
	path        string
	query       map[string]any
	body        map[string]any
	publicKey   string
	privateKey  string
	environment string
}

func getNonce() string {
	return fmt.Sprint(time.Now().UnixMilli())
}

func getSignature(privateKey string, data string, nonce string, path string) (string, error) {
	message := sha256.New()
	message.Write([]byte(nonce + data))
	return sign(privateKey, []byte(path+string(message.Sum(nil))))
}

func sign(privateKey string, message []byte) (string, error) {
	key, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	hmacHash := hmac.New(sha512.New, key)
	hmacHash.Write(message)
	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil)), nil
}

func generateRequest(c *requestParams) (*http.Request, error) {
	url := c.environment + c.path
	var queryString string
	if len(c.query) > 0 {
		queryValues, err := helpers.MapToURLValues(c.query)
		if err != nil {
			return nil, fmt.Errorf("query to URL values: %s", err)
		}
		queryString = queryValues.Encode()
		url += "?" + queryString
	}
	var nonce any
	bodyMap := c.body
	if len(c.publicKey) > 0 {
		if bodyMap == nil {
			bodyMap = make(map[string]any)
		}
		var ok bool
		nonce, ok = bodyMap["nonce"]
		if !ok {
			nonce = getNonce()
			bodyMap["nonce"] = nonce
		}
	}
	headers := make(http.Header)
	var bodyReader io.Reader
	var bodyString string
	if len(bodyMap) > 0 {
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("json marshal: %s", err)
		}
		bodyString = string(bodyBytes)
		bodyReader = bytes.NewReader(bodyBytes)
		headers.Set("Content-Type", "application/json")
	}
	request, err := http.NewRequest(c.method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("http new request: %s", err)
	}
	if len(c.publicKey) > 0 {
		signature, err := getSignature(c.privateKey, queryString+bodyString, fmt.Sprint(nonce), c.path)
		if err != nil {
			return nil, fmt.Errorf("get signature: %s", err)
		}
		headers.Set("API-Key", c.publicKey)
		headers.Set("API-Sign", signature)
		headers.Set("User-Agent", UserAgent)
	}
	request.Header = headers
	return request, nil
}

func request(rp *requestParams) (*http.Response, error) {
	req, err := generateRequest(rp)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func NewClient() (Kraken, error) {
	kClient, err := newClient()
	if err != nil {
		return nil, err
	}
	return kClient, nil
}

func newClient() (*kraken, error) {
	const apiKeyEnvName = "KRAKEN_API_KEY"
	const privateKeyEnvName = "KRAKEN_PRIVATE_KEY"
	envMap, err := helpers.LoadEnv([]string{apiKeyEnvName, privateKeyEnvName})
	if err != nil {
		return nil, fmt.Errorf("error loading api keys needed for Kraken: %w", err)
	}
	return &kraken{
		apiKey:     (*envMap)[apiKeyEnvName],
		privateKey: (*envMap)[privateKeyEnvName],
	}, nil
}
