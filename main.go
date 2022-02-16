package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Config struct {
	Proxy    Proxy     `json:"proxy"`
	Backends []Backend `json:"backends"`
}

type Proxy struct {
	Port string `json:"port"`
}

type Backend struct {
	URL    string `json:"url"`
	IsDead bool
	mu     sync.RWMutex
}

func (b *Backend) SetState(state bool) {
	b.mu.Lock()
	b.IsDead = state
	b.mu.Unlock()
}

func (b *Backend) IsDown() bool {
	b.mu.Lock()
	isAlive := b.IsDead
	b.mu.Unlock()

	return isAlive
}

var mu sync.Mutex
var idx int = 0

func lbHandler(w http.ResponseWriter, r *http.Request) {
	numBackends := len(config.Backends)

	mu.Lock()

	currentBackend := config.Backends[idx%numBackends]
	if currentBackend.IsDown() {
		idx++
	}

	targetUrl, err := url.Parse(config.Backends[idx%numBackends].URL)
	if err != nil {
		log.Fatal(err.Error())
	}
	idx++
	mu.Unlock()

	reverseProxy := httputil.NewSingleHostReverseProxy(targetUrl)
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("%v is dead.", targetUrl)
		currentBackend.SetState(true)
		lbHandler(w, r)
	}
}

func isAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)
	if err != nil {
		log.Printf("Unreachable to %v, error:", url.Host, err.Error())
		return false
	}

	defer conn.Close()
	return true
}

func healthCheck() {
	t := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-t.C:
			for _, backend := range config.Backends {
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
				log.Printf("%v checked %v by healthcheck", backend.URL, msg)
			}
		}
	}
}

var config Config

func main() {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	json.Unmarshal(data, &config)

	s := http.Server{
		Addr:    ":" + config.Proxy.Port,
		Handler: http.HandlerFunc(lbHandler),
	}

	if err = s.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
