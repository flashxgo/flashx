package flashx

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
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

	// NumberOfRequestsPerSecond states the
	// maximum number of operations to perform per second
	// If this value is not set, rate limiting will be disabled
	NumberOfRequestsPerSecond int

	// The transport used to perform proxy requests.
	// If nil, http.DefaultTransport is used.
	Transport http.RoundTripper

	// URLs is an array of string URLs that need to be configured
	URLs []string

	// LoadBalancingStrategy holds a load balancing strategy
	LoadBalancingStrategy int

	limiter ratelimit.Limiter

	proxy *httputil.ReverseProxy

	currentIndex uint64

	urls []*url.URL
}

const (
	// Nil or no strategy
	Nil int = iota

	// RoundRobin strategy
	// In this strategy, URLs will be picked
	// from the URL array one after the other
	RoundRobin
)

// Setup creates a reverse proxy for the configured URL
func (e *Engine) Setup() error {
	if e.NumberOfRequestsPerSecond > 0 {
		e.limiter = ratelimit.New(e.NumberOfRequestsPerSecond)
	} else {
		e.limiter = ratelimit.NewUnlimited()
	}

	if err := e.validateURLs(); err != nil {
		return err
	}
	return nil
}

// Initiate routes in the request,
// and routes out the response for a particular URL.
// The function accepts a response writer,
// a pointer to a request
func (e *Engine) Initiate(writer http.ResponseWriter, request *http.Request) {
	routeURL := e.getURL()

	e.limiter.Take()

	e.blacklist(writer, request)

	revProxy := httputil.NewSingleHostReverseProxy(routeURL)
	e.proxy = revProxy
	e.setupReverseProxy(routeURL)

	revProxy.ServeHTTP(writer, request)
}

func (e *Engine) validateURLs() error {
	parsedURLs := make([]*url.URL, 0)
	for _, value := range e.URLs {
		parsedURL, err := url.Parse(value)
		if err != nil {
			return err
		}
		parsedURLs = append(parsedURLs, parsedURL)
	}
	e.urls = parsedURLs
	return nil
}

func (e *Engine) getURL() *url.URL {
	if e.LoadBalancingStrategy == RoundRobin {
		nextURLIndex := int(atomic.AddUint64(&e.currentIndex, uint64(1)) % uint64(len(e.urls)))
		return e.urls[nextURLIndex]
	}
	return e.urls[0]
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

func (e *Engine) setupReverseProxy(url *url.URL) {
	e.proxy.BufferPool = e.BufferPool
	e.proxy.ErrorHandler = e.ErrorHandler
	e.proxy.ErrorLog = e.ErrorLog
	e.proxy.FlushInterval = e.FlushInterval
	e.proxy.Transport = e.Transport

	if e.ModifyRequest == nil {
		e.proxy.Director = defaultDirector(url)
	} else {
		e.proxy.Director = e.ModifyRequest
	}

	if e.ModifyResponse == nil {
		e.proxy.ModifyResponse = defaultModifyResponse()
	} else {
		e.proxy.ModifyResponse = e.ModifyResponse
	}
}

func defaultDirector(url *url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.Host = url.Host
	}
}

func defaultModifyResponse() func(*http.Response) error {
	return func(h *http.Response) error {
		return nil
	}
}
