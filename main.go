package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
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

var mu sync.Mutex
var idx int = 0

func lbHandler(w http.ResponseWriter, r *http.Request) {
	numBackends := len(config.Backends)

	mu.Lock()
	targetUrl, err := url.Parse(config.Backends[idx%numBackends].URL)
	if err != nil {
		log.Fatal(err.Error())
	}
	idx++
	mu.Unlock()

	reverseProxy := httputil.NewSingleHostReverseProxy(targetUrl)
	reverseProxy.ServeHTTP(w, r)
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
