package config

import (
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

const (
	HeartbeatIPNotContain = "HeartbeatIPNotContain"
	HeartbeatIPConflict   = "HeartbeatIPConflict"
	HeartbeatUUIDConflict = "HeartbeatUUIDConflict"
	HeartbeatUnknownError = "HeartbeatUnknownError"

	HeartbeatNotFound = "HeartbeatNotFound"
	HeartbeatSuccess  = "HeartbeatSuccess"
)

type HeartbeatResult struct {
	Result string `json:"result"`
}

type ServerInfo struct {
	Kubeconfig string
	MasterURL  string
	SvcCIDR    string
	PodCIDR    string
	TunServer  string
	//TunServerPort string
	ClientCIDR string

	//CIDR          string  // ip池
	Key           string
	Protocol      string
	WebSocketPath string
	Obfs          bool
}

type TunConfig struct {
	LocalAddr  string
	ServerAddr string
	//IntranetServerIP string
	CIDR          string
	DstCIDR       string
	Key           string
	Protocol      string
	WebSocketPath string
	ServerMode    bool
	GlobalMode    bool
	Obfs          bool
	MTU           int
	Timeout       int
	LocalGateway  string
}

func (c *TunConfig) UpdateFromEnv() {
	if cidr, exist := os.LookupEnv("K8S_TUN_CIDR"); exist {
		c.CIDR = cidr
	}
	if key, exist := os.LookupEnv("K8S_TUN_KEY"); exist {
		c.Key = key
	}
	if p, exist := os.LookupEnv("K8S_TUN_PROTOCOL"); exist {
		c.Protocol = p
	}
	if wsp, exist := os.LookupEnv("K8S_TUN_WEBSOCKET_PATH"); exist {
		c.WebSocketPath = wsp
	}
	if obfs, exist := os.LookupEnv("K8S_TUN_OBFS"); exist {
		if strings.ToLower(obfs) == "true" {
			c.Obfs = true
		} else {
			c.Obfs = false
		}
	}
	c.Validate()
}

func (c *TunConfig) Validate() bool {
	_, _, err := net.ParseCIDR(c.CIDR)
	if err != nil {
		log.WithError(err).Fatalf("c.CIDR error")
	}
	if c.Protocol == "ws" || c.Protocol == "wss" || c.Protocol == "tcp" || c.Protocol == "udp" {
	} else {
		log.Fatalf("c.Protocol: %s error", c.Protocol)
	}
	return true
}
