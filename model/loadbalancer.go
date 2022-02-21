package model

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type LoadBalancer struct {
	Conf *Config
	id   int
	mu   sync.Mutex
}

// initializes a load balancer
func Init() *LoadBalancer {
	return &LoadBalancer{
		Conf: ReadConfig(),
		id:   0,
	}
}

// load balancer main handler
func (lb *LoadBalancer) Handler(w http.ResponseWriter, r *http.Request) {
	numBackends := len(lb.Conf.Backends)

	lb.mu.Lock()
	if lb.id == numBackends {
		lb.id = 0
	}

	currentBackend := lb.Conf.Backends[lb.id%numBackends]
	if currentBackend.IsDown() {
		lb.id++
	}

	targetUrl, err := url.Parse(currentBackend.URL)
	if err != nil {
		log.Fatal(err.Error())
	}

	// reseting load balancer's backend id
	log.Printf("request incoming, redirected to %v (backend %d)\n\n", targetUrl, lb.id+1)
	lb.id++

	lb.mu.Unlock()

	reverseProxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// active healthcheck
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("%v is dead\n\n", targetUrl)
		currentBackend.SetState(true)
		lb.Handler(w, r)
	}

	reverseProxy.ServeHTTP(w, r)
}

// passive healthcheck
func (lb *LoadBalancer) HealthCheck() {
	t := time.NewTicker(time.Minute * 1)

	for {
		<-t.C
		for _, backend := range lb.Conf.Backends {
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
