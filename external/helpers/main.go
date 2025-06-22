package helpers

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

const gobDirPath = "./gobs/"

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
