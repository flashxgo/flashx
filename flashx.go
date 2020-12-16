package flashx

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Engine provides configuration options to setup and
// use FlashX
type Engine struct {
	// ModifyRequest allows you to modify the request before sending it
	// It accepts a function that alters the request to be sent.
	// Accepted function must not access the provided request after returning
	// If not set, a default value will be picked up
	ModifyRequest func(*http.Request)

	// ModifyResponse allows you to modify the response once it is received
	// It accepts a function that alters the response before returning it
	// If ModifyResponse returns an error, ErrorHandler is called
	// with its error value. If ErrorHandler is nil, its default
	// implementation is used.
	// If not set, a default value will be picked up
	ModifyResponse func(*http.Response) error
}

// Setup creates a reverse proxy for the configured URL
func (e *Engine) Setup(url *url.URL, writer http.ResponseWriter, request *http.Request) error {
	revProxy := httputil.NewSingleHostReverseProxy(url)
	if e.ModifyRequest == nil {
		e.ModifyRequest = defaultDirector(url)
	}
	revProxy.Director = e.ModifyRequest

	if e.ModifyResponse == nil {
		e.ModifyResponse = defaultModifyResponse()
	}
	revProxy.ModifyResponse = e.ModifyResponse

	revProxy.ServeHTTP(writer, request)
	return nil
}

func parseURLs(urls string) (*url.URL, error) {
	url, err := url.Parse(urls)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func defaultDirector(url *url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = url.Host
	}
}

func defaultModifyResponse() func(*http.Response) error {
	return func(h *http.Response) error {
		return nil
	}
}
