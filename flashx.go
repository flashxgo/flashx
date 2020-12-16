package flashx

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Engine provides configuration options to setup and
// use FlashX
type Engine struct {
	URLs  []string
	proxy *httputil.ReverseProxy
}

var (
	errMalformedURL = errors.New("malformed url")
)

// Setup creates a reverse proxy for the configured URL
func (e *Engine) Setup(url string, writer http.ResponseWriter, request *http.Request) error {
	endpoint, err := parseURLs(url)
	if err != nil {
		return errMalformedURL
	}
	revProxy := httputil.NewSingleHostReverseProxy(endpoint)
	e.proxy = revProxy
	return nil
}

// ModifyRequest allows you to modify the request before sending it
// It accepts a function that alters the request to be sent.
// Accepted function must not access the provided request after returning
func (e *Engine) ModifyRequest(modify func(h *http.Request)) {
	e.proxy.Director = modify
}

// ModifyResponse allows you to modify the response once it is received
// It accepts a function that alters the response before returning it
// If ModifyResponse returns an error, ErrorHandler is called
// with its error value. If ErrorHandler is nil, its default
// implementation is used.
func (e *Engine) ModifyResponse(modify func(*http.Response) error) {
	e.proxy.ModifyResponse = modify
}

func parseURLs(urls string) (*url.URL, error) {
	url, err := url.Parse(urls)
	if err != nil {
		return nil, err
	}
	return url, nil
}
