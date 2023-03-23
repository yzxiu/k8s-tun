package netutil

import (
	"net"
	"strings"
)

func FirstIP(cidr string) (string, error) {
	ips, err := Hosts(cidr)
	if err != nil {
		return "", err
	}
	return ips[0], nil
}

func CidrIps(cidr string) ([]string, error) {
	ips, err := Hosts(cidr)
	if err != nil {
		return nil, err
	}
	netmask := strings.Split(cidr, "/")[1]
	cidrips := []string{}
	for _, ip := range ips {
		cidrips = append(cidrips, ip+"/"+netmask)
	}
	return cidrips, nil
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
