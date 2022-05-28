package loadbalancer

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Provider interface {
	HandleRequest(w http.ResponseWriter, r *http.Request)
}

type loadBalancingProvider struct {
	Conf *config
	id   int
	mu   sync.Mutex
}

func NewLoadBalancingProvider() *loadBalancingProvider {
	return &loadBalancingProvider{
		Conf: ReadConfig(),
		id:   0,
	}
}

// running the load balancer
func (lb *loadBalancingProvider) ListenAndServe() {
	// creating server
	s := http.Server{
		Addr:    ":" + lb.Conf.ProxyPort,
		Handler: http.HandlerFunc(lb.HandleRequest),
	}

	// error logging
	log.Printf("load balancer starting on port " + s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("error starting load balancer: %v", err.Error())
	}
}

// load balancer main handler
func (lb *loadBalancingProvider) HandleRequest(w http.ResponseWriter, r *http.Request) {
	numBackends := len(lb.Conf.Servers)

	lb.mu.Lock()
	if lb.id == numBackends {
		lb.id = 0
	}

	currentBackend := lb.Conf.Servers[lb.id%numBackends]
	if currentBackend.IsDown() {
		lb.id++
	}

	targetUrl, err := url.Parse(currentBackend.URL)
	if err != nil {
		log.Fatal(err.Error())
	}

	// reseting load balancer's backend id
	log.Printf("request incoming, redirected to %v (%v)\n\n", targetUrl, currentBackend.Name)
	lb.id++

	lb.mu.Unlock()

	reverseProxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// active healthcheck
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("%v is dead\n\n", targetUrl)
		currentBackend.SetState(true)
		lb.HandleRequest(w, r)
	}

	reverseProxy.ServeHTTP(w, r)
}

// passive healthcheck
func (lb *loadBalancingProvider) HealthCheck() {
	t := time.NewTicker(time.Minute * 1)

	for {
		<-t.C
		for _, backend := range lb.Conf.Servers {
			pingUrl, err := url.Parse(backend.URL)
			if err != nil {
				log.Fatal(err.Error())
			}

			isAlive := isAlive(pingUrl)
			backend.SetState(!isAlive)
			msg := "ok"

			if !isAlive {
				msg = "dead"
			}
			log.Printf("%v checked %v by healthcheck\n\n", backend.URL, msg)
		}
	}
}

// backend state check
func isAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)
	if err != nil {
		log.Printf("Unreachable to %v, error: %v\n\n", url.Host, err.Error())
		return false
	}

	defer conn.Close()
	return true
}
