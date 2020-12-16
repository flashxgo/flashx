package flashx

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"go.uber.org/ratelimit"
)

// Engine provides configuration options to setup and
// use FlashX
type Engine struct {
	// BlacklistIPs is an array of IPs that needs to be blacklisted
	BlacklistIPs []string

	// A BufferPool is an interface for getting and returning temporary
	// byte slices for use by io.CopyBuffer.
	BufferPool httputil.BufferPool

	// ErrorHandler is an optional function that handles errors
	// reaching the backend or errors from ModifyResponse.
	//
	// If nil, the default is to log the provided error and return
	// a 502 Status Bad Gateway response.
	ErrorHandler func(http.ResponseWriter, *http.Request, error)

	// ErrorLog specifies an optional logger for errors
	// that occur when attempting to proxy the request.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger

	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	// If zero, no periodic flushing is done.
	// A negative value means to flush immediately
	// after each write to the client.
	// The FlushInterval is ignored when ReverseProxy
	// recognizes a response as a streaming response;
	// for such responses, writes are flushed to the client
	// immediately.
	FlushInterval time.Duration

	// NumberOfRequestsPerSecond states the
	// maximum number of operations to perform per second
	// If this value is not set, rate limiting will be disabled
	NumberOfRequestsPerSecond int

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

	// The transport used to perform proxy requests.
	// If nil, http.DefaultTransport is used.
	Transport http.RoundTripper

	limiter ratelimit.Limiter
}

// Setup creates a reverse proxy for the configured URL
func (e *Engine) Setup() {
	if e.NumberOfRequestsPerSecond > 0 {
		e.limiter = ratelimit.New(e.NumberOfRequestsPerSecond)
	} else {
		e.limiter = ratelimit.NewUnlimited()
	}
}

// Initiate routes in the request and routes out the response
func (e *Engine) Initiate(url *url.URL, writer http.ResponseWriter, request *http.Request) {
	e.limiter.Take()

	e.blacklist(writer, request)

	revProxy := httputil.NewSingleHostReverseProxy(url)
	e.setupReverseProxy(url, revProxy)

	revProxy.ServeHTTP(writer, request)
}

func (e *Engine) blacklist(writer http.ResponseWriter, request *http.Request) {
	if len(e.BlacklistIPs) > 0 {
		for _, ip := range e.BlacklistIPs {
			if ip == request.RemoteAddr {
				writer.WriteHeader(http.StatusForbidden)
				return
			}
		}
	}
}

func (e *Engine) setupReverseProxy(url *url.URL, revProxy *httputil.ReverseProxy) {
	revProxy.BufferPool = e.BufferPool
	revProxy.ErrorHandler = e.ErrorHandler
	revProxy.ErrorLog = e.ErrorLog
	revProxy.FlushInterval = e.FlushInterval
	revProxy.Transport = e.Transport

	if e.ModifyRequest == nil {
		e.ModifyRequest = defaultDirector(url)
	}
	revProxy.Director = e.ModifyRequest

	if e.ModifyResponse == nil {
		e.ModifyResponse = defaultModifyResponse()
	}
	revProxy.ModifyResponse = e.ModifyResponse
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
