package kraken

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"kasegu/external/helpers"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const BaseURL = "https://api.kraken.com"
const PublicWSURL = "wss://ws.kraken.com/v2"
const UserAgent = "KaseguKrakenClient/0.1.0"

type Kraken interface {
	Close() error
	GetAccountBalance() (*map[string]string, error)
	GetOHCLData(pair string, interval uint16) (*map[string]any, error)
	CandlesWS(method string, symbol string, interval uint16) error
}
type kraken struct {
	apiKey     string
	privateKey string
	conn       *websocket.Conn
	loopCancel *context.CancelFunc
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

func openConnection(endpoint string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to establish websocket connection to endpoint %s: %w", endpoint, err)
	}
	return c, nil
}

func wsLoop(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("closing loop")
			return
		default:
			_, p, err := conn.ReadMessage()
			if err != nil {
				return
			}
			fmt.Println("Got message:", string(p))
		}
	}
}

func New() (Kraken, error) {
	const apiKeyEnvName = "KRAKEN_API_KEY"
	const privateKeyEnvName = "KRAKEN_PRIVATE_KEY"
	envMap, err := helpers.LoadEnv([]string{apiKeyEnvName, privateKeyEnvName})
	if err != nil {
		return nil, fmt.Errorf("error loading api keys needed for Kraken: %w", err)
	}
	c, err := openConnection(PublicWSURL)
	if err != nil {
		return nil, fmt.Errorf("error opening Kraken connection: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go wsLoop(c, ctx)
	return &kraken{
		apiKey:     (*envMap)[apiKeyEnvName],
		privateKey: (*envMap)[privateKeyEnvName],
		conn:       c,
		loopCancel: &cancel,
	}, nil
}

func (k *kraken) Close() error {
	fmt.Println("closing connection gracefully")
	(*k.loopCancel)()
	err := helpers.CloseWebsocket(k.conn)
	if err != nil {
		return fmt.Errorf("error closing Kraken connection: %w", err)
	}
	fmt.Println("closed connection gracefully")
	return nil
}

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

type candlesWSParamsParams struct {
	Channel  string   `json:"channel"`
	Symbol   []string `json:"symbol"`
	Interval uint16   `json:"interval"`
}
type CandlesWSParams struct {
	Method string                `json:"method"`
	Params candlesWSParamsParams `json:"params"`
}

func (k *kraken) CandlesWS(method string, symbol string, interval uint16) error {
	var params = CandlesWSParams{
		Method: method,
		Params: candlesWSParamsParams{
			Channel:  "ohlc",
			Symbol:   []string{symbol},
			Interval: interval,
		},
	}
	err := k.conn.WriteJSON(params)
	if err != nil {
		return fmt.Errorf("error sending candles request: %w", err)
	}
	return nil
}
