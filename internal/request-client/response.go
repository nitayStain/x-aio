package requestClient

import (
	"io"
	"net/http"
)

type Response struct {
	payload string
	status  int
}

// fetching a response's data to a new readable struct
func ResponseFromHttp(res *http.Response) (*Response, error) {
	response := &Response{}
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	response.payload = string(payload)
	response.status = res.StatusCode

	return response, nil
}
