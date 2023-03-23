package tun

import (
	"net"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
)

type IptablesOp string

const IptablesOpAdd = "-A"
const IptablesOpDel = "-D"

func CreateTun(config *config.TunConfig) (iface *water.Interface) {
	c := GetConfig(config) //water.TunConfig{DeviceType: water.TUN}
	iface, err := water.New(c)
	if err != nil {
		log.Fatalln("failed to create tun interface:", err)
	}
	log.Debugf("interface created:%s", iface.Name())
	ConfigTun(config, iface)
	configClientRoute(config, iface)
	OpServerSNAT(config, IptablesOpAdd)
	return iface
}

// configure client route
func configClientRoute(config *config.TunConfig, iface *water.Interface) {
	if config.ServerMode {
		return
	}
	// Get the first ip address in the network
	gw, err := netutil.FirstIP(config.CIDR)
	if err != nil {
		log.Fatalf("Failed to get [gw] with [config.CIDR: %s]", config.CIDR)
	}
	for _, s := range strings.Split(config.DstCIDR, ",") {
		log.Debugf("config client dst cidr: %s  gw:%s", s, gw)
		ConfigClientRoute(s, gw, iface, config)
	}
}

// OpServerSNAT 服务端配置SNAT
// iptables -t nat -A POSTROUTING -s 10.99.99.0/24 -j MASQUERADE
// iptables -t nat -D POSTROUTING -s 10.99.99.0/24 -j MASQUERADE
func OpServerSNAT(config *config.TunConfig, op IptablesOp) {
	if !config.ServerMode {
		return
	}
	_, nw, err := net.ParseCIDR(config.CIDR)
	if err != nil || len(nw.String()) < 1 {
		log.Fatalf("config.CIDR error")
	}
	os := runtime.GOOS
	if os == "linux" {
		netutil.ExecCmd("/sbin/iptables", "-t", "nat", string(op), "POSTROUTING", "-s", nw.String(), "-j", "MASQUERADE")
	} else {
		log.Debugf("server snat not support os:%v", os)
	}
}

func Reset(config *config.TunConfig) {
	os := runtime.GOOS
	if os == "darwin" && !config.ServerMode && config.GlobalMode {
		netutil.ExecCmd("route", "add", "default", config.LocalGateway)
		netutil.ExecCmd("route", "change", "default", config.LocalGateway)
	}
	if os == "linux" && config.ServerMode {
		OpServerSNAT(config, IptablesOpDel)
	}
	if os == "windows" && !config.ServerMode {
		for _, s := range strings.Split(config.DstCIDR, ",") {
			_, ipnet, err := net.ParseCIDR(s)
			if err != nil {
				log.Panicf("error cidr %v", config.CIDR)
			}
			netutil.ExecCmd("route", "delete", ipnet.IP.String())
		}
	}
}

func GetMaskStr(mask net.IPMask) string {
	val := make([]byte, len(mask))
	copy(val, mask)
	var s []string
	for _, i := range val[:] {
		s = append(s, strconv.Itoa(int(i)))
	}
	return strings.Join(s, ".")
}
