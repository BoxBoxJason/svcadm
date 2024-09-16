package utils

import (
	"io"
	"net/http"
)

func GetFileContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
