package tun

import (
	"strconv"
	"strings"

	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
)

func GetConfig(config *config.TunConfig) water.Config {
	c := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "k8s-tun-0",
		},
	}
	return c
}

func ConfigTun(config *config.TunConfig, iface *water.Interface) {
	netutil.ExecCmd("/sbin/ip", "link", "set", "dev", iface.Name(), "mtu", strconv.Itoa(config.MTU))
	netutil.ExecCmd("/sbin/ip", "addr", "add", config.CIDR, "dev", iface.Name())
	netutil.ExecCmd("/sbin/ip", "link", "set", "dev", iface.Name(), "up")
	if !config.ServerMode && config.GlobalMode {
		physicalIface := netutil.GetPhysicalInterface()
		serverIP := netutil.LookupIP(strings.Split(config.ServerAddr, ":")[0])
		if physicalIface != "" && serverIP != "" {
			netutil.ExecCmd("/sbin/ip", "route", "add", "0.0.0.0/1", "dev", iface.Name())
			netutil.ExecCmd("/sbin/ip", "route", "add", "128.0.0.0/1", "dev", iface.Name())
			netutil.ExecCmd("/sbin/ip", "route", "add", "8.8.8.8/32", "via", config.LocalGateway, "dev", physicalIface)
			netutil.ExecCmd("/sbin/ip", "route", "add", strings.Join([]string{serverIP, "32"}, "/"), "via", config.LocalGateway, "dev", physicalIface)
		}
	}

	// linux服务端，开启转发
	// echo "1" > /proc/sys/net/ipv4/ip_forward
	if config.ServerMode {
		netutil.ExecCmd("/sbin/sysctl", "-p")
	}
}

func ConfigClientRoute(dst string, gw string, iface *water.Interface, onfig *config.TunConfig) {
	netutil.ExecCmd("/sbin/ip", "route", "add", dst, "via", gw)
}
