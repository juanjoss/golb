package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fuato1/golb/model"
)

func runBackend(port string, id int) {
	// creating server
	s := http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "backend %d\n", id)
		}),
	}
	fmt.Println("backend ", id, " running on port ", port)

	// error logging
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}

	os.Exit(1)
}

func main() {
	// initializing load balancer
	lb := model.Init()

	// creating backends
	go runBackend("3001", 1)
	go runBackend("3002", 2)

	// running passive healthcheck routine
	go lb.HealthCheck()

	// creating server
	s := http.Server{
		Addr:    ":" + lb.Conf.ProxyPort,
		Handler: http.HandlerFunc(lb.Handler),
	}

	// error logging
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
