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
	"strings"
	"time"
)

const BaseURL = "https://api.kraken.com"
const UserAgent = "KaseguKrakenClient/0.1.0"

type Kraken interface {
	GetAccountBalance() (*map[string]string, error)
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

/*func getKrakenSignature(urlPath string, data interface{}, secret string) (string, error) {
	var encodedData string
	switch v := data.(type) {
	case string:
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(v), &jsonData); err != nil {
			return "", err
		}
		encodedData = jsonData["nonce"].(string) + v
	case map[string]interface{}:
		dataMap := urlMod.Values{}
		for key, value := range v {
			dataMap.Set(key, fmt.Sprintf("%v", value))
		}
		encodedData = v["nonce"].(string) + dataMap.Encode()
	default:
		return "", fmt.Errorf("invalid data type")
	}
	fmt.Println(encodedData)
	sha := sha256.New()
	sha.Write([]byte(encodedData))
	shaSum := sha.Sum(nil)
	message := append([]byte(urlPath), shaSum...)
	d, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decode failed: %w", err)
	}
	mac := hmac.New(sha512.New, d)
	mac.Write(message)
	macSum := mac.Sum(nil)
	sigDigest := base64.StdEncoding.EncodeToString(macSum)
	return sigDigest, nil
}*/

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
			/*t, err := strconv.ParseInt(getNonce(), 10, 64)
			if err != nil {
				return nil, err
			}
			nonce = t*/
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
		//fmt.Println(bodyString)
		headers.Set("Content-Type", "application/json")
	}
	request, err := http.NewRequest(c.method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("http new request: %s", err)
	}
	if len(c.publicKey) > 0 {
		signature, err := getSignature(c.privateKey, queryString+bodyString, fmt.Sprint(nonce), c.path)
		//signature, err := getKrakenSignature(c.path, bodyMap, c.privateKey)
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
	//fmt.Println(req)
	return http.DefaultClient.Do(req)
}

func New() (Kraken, error) {
	const apiKeyEnvName = "KRAKEN_API_KEY"
	const privateKeyEnvName = "KRAKEN_PRIVATE_KEY"
	envMap, err := helpers.LoadEnv([]string{apiKeyEnvName, privateKeyEnvName})
	if err != nil {
		return nil, fmt.Errorf("error loading api keys needed for Kraken: %w", err)
	}
	return &kraken{apiKey: (*envMap)[apiKeyEnvName], privateKey: (*envMap)[privateKeyEnvName]}, nil
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
		return nil, fmt.Errorf("error getting account balance: %w", err)
	}
	defer helpers.CheckedClose(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error getting account balance: %w", err)
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
