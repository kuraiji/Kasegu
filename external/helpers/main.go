package helpers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	urlMod "net/url"
	"os"
	"reflect"

	_ "github.com/joho/godotenv/autoload"
)

const gobDirPath = "./gobs/"

type RequestParams struct {
	Method      string
	Path        string
	Query       *map[string]any
	Body        *map[string]any
	Header      *map[string]string
	Environment string
}

func Request(rp *RequestParams) (*http.Response, error) {
	url := rp.Environment + rp.Path
	if rp.Query != nil && len(*rp.Query) > 0 {
		qv, err := MapToURLValues(*rp.Query)
		if err != nil {
			return nil, fmt.Errorf("error converting query params to URL values: %w", err)
		}
		url += "?" + qv.Encode()
	}
	var br io.Reader
	if rp.Body != nil && len(*rp.Body) > 0 {
		bb, err := json.Marshal(rp.Body)
		if err != nil {
			return nil, fmt.Errorf("error converting body to JSON: %w", err)
		}
		br = bytes.NewReader(bb)
	}
	req, err := http.NewRequest(rp.Method, url, br)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	var h http.Header
	if rp.Header != nil && len(*rp.Header) > 0 {
		for k, v := range *rp.Header {
			h[k] = []string{v}
		}
	}
	req.Header = h
	return http.DefaultClient.Do(req)
}

func CheckedClose(Body io.ReadCloser) {
	err := Body.Close()
	if err != nil {
		log.Println(fmt.Errorf("failed closing file: %w", err))
	}
}

func LoadEnv(envNames []string) (*map[string]string, error) {
	envMap := make(map[string]string)
	for _, envName := range envNames {
		envValue, ok := os.LookupEnv(envName)
		if !ok {
			return nil, fmt.Errorf("environment variable %s not found", envName)
		}
		envMap[envName] = envValue
	}
	return &envMap, nil
}

func GetWithHeaders(client *http.Client, url string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating new request: %w", err)
	}
	req.Header = headers
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to endpoint: %w", err)
	}
	return resp, nil
}

func IsThereSerializedData(filename string) bool {
	url := fmt.Sprintf("%s%s", gobDirPath, filename)
	if _, err := os.Stat(url); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func SerializeData[T interface{}](data *T, filename string) error {
	url := fmt.Sprintf("%s%s", gobDirPath, filename)
	if _, err := os.Stat(gobDirPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(gobDirPath, 0755)
		if err != nil {
			return fmt.Errorf("failed creating folder: %w", err)
		}
	}
	file, err := os.Create(url)
	if err != nil {
		return fmt.Errorf("failed creating file: %w", err)
	}
	defer CheckedClose(file)
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(*data)
	if err != nil {
		return fmt.Errorf("failed encoding: %w", err)
	}
	return nil
}

func UnserializeData[T interface{}](filename string) (*T, error) {
	url := fmt.Sprintf("%s%s", gobDirPath, filename)
	file, err := os.Open(url)
	if err != nil {
		return nil, fmt.Errorf("failed opening file: %w", err)
	}
	defer CheckedClose(file)
	decoder := gob.NewDecoder(file)
	var data *T
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed decoding: %w", err)
	}
	return data, nil
}

func AppendQueryParameters(url string, params *map[string]string) string {
	if len(*params) == 0 {
		return url
	}
	queryParams := make(urlMod.Values)
	for key, value := range *params {
		queryParams.Add(key, value)
	}
	return url + "?" + queryParams.Encode()
}

func MapToURLValues(m map[string]any) (urlMod.Values, error) {
	uv := urlMod.Values{}
	for k, v := range m {
		switch v := v.(type) {
		case []string:
			uv[k] = v
		case string:
			uv[k] = []string{v}
		default:
			j, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			uv[k] = []string{string(j)}
		}
	}
	return uv, nil
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
