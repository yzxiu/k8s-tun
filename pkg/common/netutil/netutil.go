package netutil

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
)

func ConnectServer(config *config.TunConfig) net.Conn {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	scheme := "ws"
	if config.Protocol == "wss" {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: config.ServerAddr, Path: config.WebSocketPath}
	header := make(http.Header)
	//header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	header.Set("key", config.Key)
	dialer := ws.Dialer{
		Header:  ws.HandshakeHeaderHTTP(header),
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	c, _, _, err := dialer.Dial(context.Background(), u.String())
	if err != nil {
		log.Warnf("[client] failed to dial websocket %s %v", u.String(), err)
		return nil
	}
	return c
}

func GetPhysicalInterface() (name string) {
	ifaces := getAllPhysicalInterfaces()
	if len(ifaces) == 0 {
		return ""
	}
	netAddrs, _ := ifaces[0].Addrs()
	for _, addr := range netAddrs {
		ip, ok := addr.(*net.IPNet)
		if ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
			name = ifaces[0].Name
			break
		}
	}
	return name
}

func getAllPhysicalInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error(err)
		return nil
	}

	var outInterfaces []net.Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp == 1 && isPhysicalInterface(iface.Name) {
			netAddrs, _ := iface.Addrs()
			if len(netAddrs) > 0 {
				outInterfaces = append(outInterfaces, iface)
			}
		}
	}
	return outInterfaces
}

func isPhysicalInterface(addr string) bool {
	prefixArray := []string{"ens", "enp", "enx", "eno", "eth", "en0", "wlan", "wlp", "wlo", "wlx", "wifi0", "lan0"}
	for _, pref := range prefixArray {
		if strings.HasPrefix(strings.ToLower(addr), pref) {
			return true
		}
	}
	return false
}

func LookupIP(domain string) string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		log.Error(err)
		return ""
	}
	for _, ip := range ips {
		return ip.To4().String()
	}
	return ""
}

func IsIPv4(packet []byte) bool {
	flag := packet[0] >> 4
	return flag == 4
}

func IsIPv6(packet []byte) bool {
	flag := packet[0] >> 4
	return flag == 6
}

func GetIPv4Source(packet []byte) net.IP {
	return net.IPv4(packet[12], packet[13], packet[14], packet[15])
}

func GetIPv4Destination(packet []byte) net.IP {
	return net.IPv4(packet[16], packet[17], packet[18], packet[19])
}

func GetIPv6Source(packet []byte) net.IP {
	return net.IP(packet[8:24])
}

func GetIPv6Destination(packet []byte) net.IP {
	return net.IP(packet[24:40])
}

func GetSourceKey(packet []byte) string {
	key := ""
	if IsIPv4(packet) && len(packet) >= 20 {
		key = GetIPv4Source(packet).To4().String()
	} else if IsIPv6(packet) && len(packet) >= 40 {
		key = GetIPv6Source(packet).To16().String()
	}
	return key
}

func GetDestinationKey(packet []byte) string {
	key := ""
	if IsIPv4(packet) && len(packet) >= 20 {
		key = GetIPv4Destination(packet).To4().String()
	} else if IsIPv6(packet) && len(packet) >= 40 {
		key = GetIPv6Destination(packet).To16().String()
	}
	return key
}

func ExecCmd(c string, args ...string) string {
	log.Debugf("%v %v", c, args)
	cmd := exec.Command(c, args...)
	out, err := cmd.Output()
	if err != nil {
		log.WithError(err).Errorf("failed to exec cmd: %v", err)
	}
	if len(out) == 0 {
		return ""
	}
	s := string(out)
	return strings.ReplaceAll(s, "\n", "")
}

func GetLinuxLocalGateway() string {
	return ExecCmd("sh", "-c", "route -n | grep 'UG[ \t]' | awk '{print $2}'")
}

func GetMacLocalGateway() string {
	return ExecCmd("sh", "-c", "route -n get default | grep 'gateway' | awk '{print $2}'")
}
