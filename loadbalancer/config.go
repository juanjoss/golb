package loadbalancer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// load balancer configuration
type config struct {
	ProxyPort string    `json:"proxyPort"`
	Servers   []*server `json:"servers"`
}

func ReadConfig() *config {
	// reading config from JSON file
	var config config
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("error reading load balancer config: %v", err.Error())
	}
	json.Unmarshal(data, &config)
	fmt.Println(config)

	return &config
}
