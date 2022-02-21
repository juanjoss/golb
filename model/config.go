package model

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// load balancer configuration
type Config struct {
	ProxyPort string     `json:"proxyPort"`
	Backends  []*Backend `json:"backends"`
}

func ReadConfig() *Config {
	// reading config from JSON file
	var config Config
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	json.Unmarshal(data, &config)

	return &config
}
