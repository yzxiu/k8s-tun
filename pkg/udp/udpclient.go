package udp

import (
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/cipher"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/tun"
)

// Start udp client
func StartClient(config *config.TunConfig) {
	log.Debugf("vtun udp client started on %v", config.LocalAddr)
	iface := tun.CreateTun(config)
	serverAddr, err := net.ResolveUDPAddr("udp", config.ServerAddr)
	if err != nil {
		log.Fatalln("failed to resolve server addr:", err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", config.LocalAddr)
	if err != nil {
		log.Fatalln("failed to get udp socket:", err)
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("failed to listen on udp socket:", err)
	}
	defer conn.Close()
	go udpToTun(config, conn, iface)
	tunToUdp(config, conn, serverAddr, iface)
}

func udpToTun(config *config.TunConfig, conn *net.UDPConn, iface *water.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, _, err := conn.ReadFromUDP(packet)
		if err != nil || n == 0 {
			continue
		}
		var b []byte
		if config.Obfs {
			b = cipher.XOR(packet[:n])
		} else {
			b = packet[:n]
		}
		iface.Write(b)
	}
}
func tunToUdp(config *config.TunConfig, conn *net.UDPConn, serverAddr *net.UDPAddr, iface *water.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		var b []byte
		if config.Obfs {
			b = cipher.XOR(packet[:n])
		} else {
			b = packet[:n]
		}
		conn.WriteToUDP(b, serverAddr)
	}
}
