package app

import (
	"encoding/json"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/yzxiu/k8s-tun/pkg/common/cipher"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"github.com/yzxiu/k8s-tun/pkg/tcp"
	"github.com/yzxiu/k8s-tun/pkg/tun"
	"github.com/yzxiu/k8s-tun/pkg/udp"
	"github.com/yzxiu/k8s-tun/pkg/ws"
)

func StartTun(tunConfig *config.TunConfig, quit <-chan struct{}) {
	initConfig(tunConfig)
	go startApp(tunConfig)
	<-quit
	stopApp(tunConfig)
}

func initConfig(config *config.TunConfig) {
	if !config.ServerMode {
	}
	if !config.ServerMode && config.GlobalMode {
		switch runtime.GOOS {
		case "linux":
			config.LocalGateway = netutil.GetLinuxLocalGateway()
		case "darwin":
			config.LocalGateway = netutil.GetMacLocalGateway()
		}
	}
	cipher.GenerateKey(config.Key)
	configJson, _ := json.Marshal(config)
	log.Debugf("init config:%s", string(configJson))
}

func startApp(config *config.TunConfig) {
	switch config.Protocol {
	case "udp":
		if config.ServerMode {
			udp.StartServer(config)
		} else {
			udp.StartClient(config)
		}
	case "tcp":
		if config.ServerMode {
			tcp.StartServer(config)
		} else {
			tcp.StartClient(config)
		}
	case "ws":
		if config.ServerMode {
			ws.StartServer(config)
		} else {
			ws.StartClient(config)
		}
	default:
		if config.ServerMode {
			ws.StartServer(config)
		} else {
			ws.StartClient(config)
		}
	}
}

func stopApp(config *config.TunConfig) {
	tun.Reset(config)
	log.Debugf("stopped!!!")
}
