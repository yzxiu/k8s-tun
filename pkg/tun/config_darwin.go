package tun

import (
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"log"
	"net"
	"strings"
)

func GetConfig(config *config.TunConfig) water.Config {
	c := water.Config{DeviceType: water.TUN}
	return c
}

func ConfigTun(config *config.TunConfig, iface *water.Interface) {

	ip, _, err := net.ParseCIDR(config.CIDR)
	if err != nil {
		log.Panicf("error cidr %v", config.CIDR)
	}

	gateway, _ := netutil.FirstIP(config.CIDR) // config.IntranetServerIP
	netutil.ExecCmd("ifconfig", iface.Name(), "inet", ip.String(), gateway, "up")
	netutil.ExecCmd("route", "add", config.CIDR, "-interface", iface.Name())

	if !config.ServerMode && config.GlobalMode {
		physicalIface := netutil.GetPhysicalInterface()
		serverIP := netutil.LookupIP(strings.Split(config.ServerAddr, ":")[0])
		if physicalIface != "" && serverIP != "" {
			netutil.ExecCmd("route", "add", serverIP, config.LocalGateway)
			netutil.ExecCmd("route", "add", "8.8.8.8", config.LocalGateway)
			netutil.ExecCmd("route", "add", "0.0.0.0/1", "-interface", iface.Name())
			netutil.ExecCmd("route", "add", "128.0.0.0/1", "-interface", iface.Name())
			netutil.ExecCmd("route", "add", "default", gateway)
			netutil.ExecCmd("route", "change", "default", gateway)
		}
	}
}

// ConfigClientRoute macos route
// route -n add -net 10.233.0.0 -netmask 255.255.0.0 10.99.99.1
func ConfigClientRoute(dst string, gw string, iface *water.Interface, config *config.TunConfig) {
	_, ipnet, err := net.ParseCIDR(dst)
	if err != nil {
		log.Panicf("error cidr %v", config.CIDR)
	}
	netutil.ExecCmd("/sbin/route", "-n", "add", "-net", ipnet.IP.String(), "-netmask", GetMaskStr(ipnet.Mask), gw)
}
