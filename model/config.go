package model

// load balancer configuration
type Config struct {
	ProxyPort string     `json:"proxyPort"`
	Backends  []*Backend `json:"backends"`
}
