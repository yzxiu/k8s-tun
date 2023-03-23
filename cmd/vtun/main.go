package main

import (
	"flag"

	"github.com/yzxiu/k8s-tun/cmd/vtun/app"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/signal"
)

func main() {
	tunConfig := config.TunConfig{}
	flag.StringVar(&tunConfig.CIDR, "c", "172.16.0.10/24", "tun interface cidr")
	flag.StringVar(&tunConfig.DstCIDR, "dst", "10.233.64.0/18,10.233.0.0/18", "vpn cidr")
	flag.IntVar(&tunConfig.MTU, "mtu", 1500, "tun mtu")
	flag.StringVar(&tunConfig.LocalAddr, "l", ":3000", "local address")
	flag.StringVar(&tunConfig.ServerAddr, "s", ":3001", "server address")
	flag.StringVar(&tunConfig.Key, "k", "freedom@2022", "key")
	flag.StringVar(&tunConfig.Protocol, "p", "wss", "protocol tcp/udp/ws/wss")
	flag.StringVar(&tunConfig.WebSocketPath, "path", "/freedom", "websocket path")
	flag.BoolVar(&tunConfig.ServerMode, "S", false, "server mode")
	flag.BoolVar(&tunConfig.GlobalMode, "g", false, "client global mode")
	flag.BoolVar(&tunConfig.Obfs, "obfs", false, "enable data obfuscation")
	flag.IntVar(&tunConfig.Timeout, "t", 30, "dial timeout in seconds")
	flag.Parse()
	tunConfig.UpdateFromEnv()
	app.StartTun(&tunConfig, signal.SetupSignalHandler())
}
