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
	urlMod "net/url"
	"time"
)

const BaseURL = "https://api.kraken.com"

type Kraken interface {
	GetAccountBalance()
}
type kraken struct {
	apiKey     string
	privateKey string
}

func getNonce() string {
	return fmt.Sprint(time.Now().UnixMilli())
}

/*func getSignature(privateKey string, data string, nonce string, path string) (string, error) {
	message := sha256.New()
	message.Write([]byte(nonce + data))
	return sign(privateKey, []byte(path+string(message.Sum(nil))))
}

func sign(privateKey string, message []byte) (string, error) {
	key, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	hmacHash := hmac.New(sha256.New, key)
	hmacHash.Write(message)
	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil)), nil
}*/

func getKrakenSignature(urlPath string, data interface{}, secret string) (string, error) {
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
	sha := sha256.New()
	sha.Write([]byte(encodedData))
	shaSum := sha.Sum(nil)
	message := append([]byte(urlPath), shaSum...)
	d, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha512.New, d)
	mac.Write(message)
	macSum := mac.Sum(nil)
	sigDigest := base64.StdEncoding.EncodeToString(macSum)
	return sigDigest, nil
}

type requestParams struct {
	method    string
	path      string
	query     *map[string]any
	body      *map[string]any
	isPrivate bool
}

func (k *kraken) generateRequestObject(rp *requestParams) (*http.Request, error) {
	url := BaseURL + rp.path
	var qs string
	if rp.query != nil && len(*rp.query) > 0 {
		qv, err := helpers.MapToURLValues(*rp.query)
		if err != nil {
			return nil, fmt.Errorf("error converting query params to URL values: %w", err)
		}
		qs = qv.Encode()
		url += "?" + qs
	}
	var nonce any
	bm := rp.body
	if bm == nil {
		bm = new(map[string]any)
	}
	if rp.isPrivate {
		if *bm == nil {
			*bm = make(map[string]any)
		}
		var ok bool
		nonce, ok = (*bm)["nonce"]
		if !ok {
			nonce = getNonce()
			(*bm)["nonce"] = nonce
		}
	}
	h := make(http.Header)
	var br io.Reader
	//var bs string
	if bm != nil && len(*bm) > 0 {
		bb, err := json.Marshal(*bm)
		if err != nil {
			return nil, fmt.Errorf("error converting body to JSON: %w", err)
		}
		//bs = string(bb)
		br = bytes.NewReader(bb)
		h.Set("Content-Type", "application/json")
	}
	req, err := http.NewRequest(rp.method, url, br)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	if rp.isPrivate {
		s, err := getKrakenSignature(rp.path, *bm, k.privateKey)
		if err != nil {
			return nil, fmt.Errorf("error generating signature: %w", err)
		}
		h.Set("API-Key", k.apiKey)
		h.Set("API-Sign", s)
	}
	req.Header = h
	return req, nil
}

func (k *kraken) request(rp *requestParams) (*http.Response, error) {
	req, err := k.generateRequestObject(rp)
	if err != nil {
		return nil, err
	}
	fmt.Println(req)
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

func (k *kraken) GetAccountBalance() {
	resp, err := k.request(&requestParams{
		method:    "POST",
		path:      "/0/private/Balance",
		isPrivate: true,
	})
	if err != nil {
		fmt.Printf("error getting account balance: %v\n", err)
	}
	defer helpers.CheckedClose(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response body: %v\n", err)
	}
	fmt.Println(string(data))
}
