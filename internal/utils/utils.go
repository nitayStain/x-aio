package utils

import (
	"io"
	"net/http"
	"time"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"

func GetPageContent(url string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{ // Force HTTP2, so the request will pass
			ForceAttemptHTTP2: true,
		},
	}

	// create a custom request, set a user-agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}
