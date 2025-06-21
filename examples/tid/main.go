package main

import (
	"net/http"
)

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
		},
	}
}
