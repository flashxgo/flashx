package flashx

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"testing"
	"time"

	"go.uber.org/ratelimit"
)

func TestEngine_Setup(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid URLS : no error",
			fields: fields{
				URLs: []string{"http://localhost:3000", "http://localhost:4000"},
			},
			wantErr: false,
		},
		{
			name: "invalid first URL : throws error",
			fields: fields{
				URLs: []string{" http://foo.com", "http://localhost:4000"},
			},
			wantErr: true,
		},
		{
			name: "invalid second URL : throws error",
			fields: fields{
				URLs: []string{"http://foo.com", "http://[fe80::%31]/"},
			},
			wantErr: true,
		},
		{
			name:    "no URLs passed : no error",
			fields:  fields{},
			wantErr: false,
		},
		{
			name: "rate limit > 0 : no error",
			fields: fields{
				URLs:                      []string{"http://foo.com", "http://localhost:4000"},
				NumberOfRequestsPerSecond: 50,
			},
			wantErr: false,
		},
		{
			name: "least connections load balancing strategy with urls : no error",
			fields: fields{
				URLs:                  []string{"http://localhost:4000"},
				LoadBalancingStrategy: LeastConnections,
			},
			wantErr: false,
		},
		{
			name: "load balancing strategy without urls: throws error",
			fields: fields{
				LoadBalancingStrategy: RoundRobin,
			},
			wantErr: true,
		},
		{
			name: "weighted round robin load balancing strategy : no error",
			fields: fields{
				URLs:                  []string{"http://localhost:3000"},
				LoadBalancingStrategy: WeightedRoundRobin,
				RoundRobinWeights:     []int{1},
			},
			wantErr: false,
		},
		{
			name: "weighted round robin load balancing strategy, no round robin weights : throws error",
			fields: fields{
				URLs:                  []string{"http://localhost:3000"},
				LoadBalancingStrategy: WeightedRoundRobin,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			if err := e.Setup(); (err != nil) != tt.wantErr {
				t.Errorf("Engine.Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_validateURLs(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid URLS : no error",
			fields: fields{
				URLs: []string{"http://localhost:3000", "http://localhost:4000"},
			},
			wantErr: false,
		},
		{
			name: "invalid first URL : throws error",
			fields: fields{
				URLs: []string{" http://foo.com", "http://localhost:4000"},
			},
			wantErr: true,
		},
		{
			name: "invalid second URL : throws error",
			fields: fields{
				URLs: []string{"http://foo.com", "http://[fe80::%31]/"},
			},
			wantErr: true,
		},
		{
			name:    "no urls : no error",
			fields:  fields{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			if err := e.validateURLs(); (err != nil) != tt.wantErr {
				t.Errorf("Engine.validateURLs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_setupReverseProxy(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}
	type args struct {
		url *url.URL
	}
	validURL := &url.URL{
		Scheme: "http",
		Host:   "localhost:3000",
	}
	modifyRequestFunc := func(req *http.Request) {
		req.URL.Host = validURL.Host
		req.URL.Scheme = validURL.Scheme
		req.Host = validURL.Host
	}
	modifyResponseFunc := func(*http.Response) error {
		return nil
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "modify request is nil",
			fields: fields{
				proxy:         &httputil.ReverseProxy{},
				ModifyRequest: nil,
			},
			args: args{
				url: validURL,
			},
		},
		{
			name: "modify response is nil",
			fields: fields{
				proxy:          &httputil.ReverseProxy{},
				ModifyResponse: nil,
			},
			args: args{
				url: validURL,
			},
		},
		{
			name: "modify request is not nil",
			fields: fields{
				proxy:         &httputil.ReverseProxy{},
				ModifyRequest: modifyRequestFunc,
			},
			args: args{
				url: validURL,
			},
		},
		{
			name: "modify response is not nil",
			fields: fields{
				proxy:          &httputil.ReverseProxy{},
				ModifyResponse: modifyResponseFunc,
			},
			args: args{
				url: validURL,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			e.setupReverseProxy(tt.args.url)
		})
	}
}

func TestEngine_blacklist(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}
	type args struct {
		writer  http.ResponseWriter
		request *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "blacklist IPs: allowed IP",
			fields: fields{
				BlacklistIPs: []string{"127.0.0.1"},
			},
			args: args{
				writer: httptest.NewRecorder(),
				request: &http.Request{
					RemoteAddr: "192.168.1.7",
				},
			},
		},
		{
			name: "blacklist IPs: forbidden IP",
			fields: fields{
				BlacklistIPs: []string{"192.168.1.7"},
			},
			args: args{
				writer: httptest.NewRecorder(),
				request: &http.Request{
					RemoteAddr: "192.168.1.7",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			e.blacklist(tt.args.writer, tt.args.request)
		})
	}
}

func TestEngine_getURL(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}

	leastConnectionsMap := make(map[*url.URL]int)
	leastConnectionsMap[&url.URL{
		Scheme: "http",
		Host:   "localhost:3000",
	}] = 1
	leastConnectionsMap[&url.URL{
		Scheme: "http",
		Host:   "localhost:4000",
	}] = 2

	tests := []struct {
		name   string
		fields fields
		want   *url.URL
	}{
		{
			name: "round robin load balancing",
			fields: fields{
				currentIndex: -1,
				urls: []*url.URL{
					{
						Scheme: "http",
						Host:   "localhost:3000",
					},
					{
						Scheme: "http",
						Host:   "localhost:4000",
					},
				},
				LoadBalancingStrategy: RoundRobin,
			},
			want: &url.URL{
				Scheme: "http",
				Host:   "localhost:3000",
			},
		},
		{
			name: "weighted round robin load balancing",
			fields: fields{
				currentIndex: -1,
				weightedURLs: []*url.URL{
					{
						Scheme: "http",
						Host:   "localhost:3000",
					},
					{
						Scheme: "http",
						Host:   "localhost:4000",
					},
				},
				LoadBalancingStrategy: WeightedRoundRobin,
			},
			want: &url.URL{
				Scheme: "http",
				Host:   "localhost:3000",
			},
		},
		{
			name: "least connections load balancing",
			fields: fields{
				urls: []*url.URL{
					{
						Scheme: "http",
						Host:   "localhost:3000",
					},
					{
						Scheme: "http",
						Host:   "localhost:4000",
					},
				},
				LoadBalancingStrategy: LeastConnections,
				leastConnectionMap:    leastConnectionsMap,
			},
			want: &url.URL{
				Scheme: "http",
				Host:   "localhost:3000",
			},
		},
		{
			name: "nil load balancer",
			fields: fields{
				urls: []*url.URL{
					{
						Scheme: "http",
						Host:   "localhost:3000",
					},
					{
						Scheme: "http",
						Host:   "localhost:4000",
					},
				},
			},
			want: &url.URL{
				Scheme: "http",
				Host:   "localhost:3000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			if got := e.getURL(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Engine.getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngine_Initiate(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}
	type args struct {
		writer  http.ResponseWriter
		request *http.Request
	}

	//backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write([]byte("I am the backend"))
		}
	}))
	defer backend.Close()
	backendURL, _ := url.Parse(backend.URL)

	// frontend
	proxyHandler := httputil.NewSingleHostReverseProxy(backendURL)
	frontend := httptest.NewServer(proxyHandler)
	defer frontend.Close()
	getReq, _ := http.NewRequest("GET", frontend.URL, nil)
	w := httptest.NewRecorder()
	parsedFrontendURL, _ := url.Parse(frontend.URL)

	leastConnectionsMap := make(map[*url.URL]int)
	leastConnectionsMap[parsedFrontendURL] = 1

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "initiate reverse proxy, no load balancer",
			fields: fields{
				urls:    []*url.URL{parsedFrontendURL},
				limiter: ratelimit.NewUnlimited(),
			},
			args: args{
				writer:  w,
				request: getReq,
			},
		},
		{
			name: "initiate reverse proxy, least connections load balancer",
			fields: fields{
				urls:                  []*url.URL{parsedFrontendURL},
				limiter:               ratelimit.NewUnlimited(),
				leastConnectionMap:    leastConnectionsMap,
				LoadBalancingStrategy: LeastConnections,
			},
			args: args{
				writer:  w,
				request: getReq,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			e.Initiate(tt.args.writer, tt.args.request)
		})
	}
}

func TestEngine_InitiateOverride(t *testing.T) {
	type fields struct {
		BlacklistIPs              []string
		BufferPool                httputil.BufferPool
		ErrorHandler              func(http.ResponseWriter, *http.Request, error)
		ErrorLog                  *log.Logger
		FlushInterval             time.Duration
		ModifyRequest             func(*http.Request)
		ModifyResponse            func(*http.Response) error
		NumberOfRequestsPerSecond int
		Transport                 http.RoundTripper
		URLs                      []string
		LoadBalancingStrategy     int
		RoundRobinWeights         []int
		limiter                   ratelimit.Limiter
		proxy                     *httputil.ReverseProxy
		currentIndex              int64
		urls                      []*url.URL
		weightedURLs              []*url.URL
		leastConnectionMap        map[*url.URL]int
	}
	type args struct {
		writer   http.ResponseWriter
		request  *http.Request
		routeURL *url.URL
	}

	//backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write([]byte("I am the backend"))
		}
	}))
	defer backend.Close()
	backendURL, _ := url.Parse(backend.URL)

	// frontend
	proxyHandler := httputil.NewSingleHostReverseProxy(backendURL)
	frontend := httptest.NewServer(proxyHandler)
	defer frontend.Close()
	getReq, _ := http.NewRequest("GET", frontend.URL, nil)
	w := httptest.NewRecorder()
	parsedFrontendURL, _ := url.Parse(frontend.URL)

	leastConnectionsMap := make(map[*url.URL]int)
	leastConnectionsMap[parsedFrontendURL] = 1

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "initiate override",
			fields: fields{
				limiter: ratelimit.NewUnlimited(),
			},
			args: args{
				writer:   w,
				request:  getReq,
				routeURL: parsedFrontendURL,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				BlacklistIPs:              tt.fields.BlacklistIPs,
				BufferPool:                tt.fields.BufferPool,
				ErrorHandler:              tt.fields.ErrorHandler,
				ErrorLog:                  tt.fields.ErrorLog,
				FlushInterval:             tt.fields.FlushInterval,
				ModifyRequest:             tt.fields.ModifyRequest,
				ModifyResponse:            tt.fields.ModifyResponse,
				NumberOfRequestsPerSecond: tt.fields.NumberOfRequestsPerSecond,
				Transport:                 tt.fields.Transport,
				URLs:                      tt.fields.URLs,
				LoadBalancingStrategy:     tt.fields.LoadBalancingStrategy,
				RoundRobinWeights:         tt.fields.RoundRobinWeights,
				limiter:                   tt.fields.limiter,
				proxy:                     tt.fields.proxy,
				currentIndex:              tt.fields.currentIndex,
				urls:                      tt.fields.urls,
				weightedURLs:              tt.fields.weightedURLs,
				leastConnectionMap:        tt.fields.leastConnectionMap,
			}
			e.InitiateOverride(tt.args.writer, tt.args.request, tt.args.routeURL)
		})
	}
}
