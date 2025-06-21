package requestClient

import (
	"net/http"
	"time"
)

// TODO: maybe implement some helper function.
// if wont, remove this type and keep it as a map
type Cookies map[string]string

type RequestClient struct {
	Client  *http.Client // Http client, builtin
	Headers *http.Header // Http headers that are being set
	Cookies Cookies      // when running a request, frankly we will be adding each cookie to the jar manually
}

// initiates a new request client with the given headers
func NewClient(userAgent string, headers map[string]string, cookies map[string]string) *RequestClient {
	rc := &RequestClient{}

	rc.Client = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
		},
	}

	rc.Cookies = cookies

	rc.Headers = &http.Header{}
	for k, v := range headers {
		rc.Headers.Set(k, v)
	}

	if userAgent != "" {
		rc.Headers.Set("User-Agent", userAgent)
	}

	return rc
}

func (c *RequestClient) MakeRequest(
	method, url string, /* TODO: add fields for payload, and another additional data */
) (*Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = *c.Headers

	// update current client's cookies for the upcoming request
	for k, v := range c.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return ResponseFromHttp(res)
}
