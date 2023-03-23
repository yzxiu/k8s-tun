package tun

import (
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"net"
	"os/exec"
)

func GetConfig(config *config.TunConfig) water.Config {
	c := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			ComponentID: "tap0901",
			Network:     config.CIDR,
		},
	}
	return c
}

func ConfigTun(config *config.TunConfig, iface *water.Interface) {
	ip, ipnet, err := net.ParseCIDR(config.CIDR)
	if err != nil {
		log.Panicf("error cidr %v", config.CIDR)
	}
	log.Debugf("windows tun interface name:%v", iface.Name())
	netutil.ExecCmd("netsh", "interface", "ip", "set", "address", iface.Name(), "static", ip.String(), GetMaskStr(ipnet.Mask), "none")

	for i := 0; i < 10; i++ {
		if exec.Command("ping", ip.String(), "-n", "1", "-w", "2000").Run() == nil {
			return
		}
	}
	log.Fatalf("init failed, plesae restart %v", err)
}

func ConfigClientRoute(dst string, gw string, iface *water.Interface, config *config.TunConfig) {
	_, ipnet, err := net.ParseCIDR(dst)
	if err != nil {
		log.Panicf("error cidr %v", config.CIDR)
	}
	netutil.ExecCmd("route", "add", ipnet.IP.String(), "mask", GetMaskStr(ipnet.Mask), gw)
}
