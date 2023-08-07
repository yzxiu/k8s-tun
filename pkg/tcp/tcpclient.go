package tcp

import (
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/cache"
	"github.com/yzxiu/k8s-tun/pkg/common/cipher"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/tun"
)

// Start tcp client
func StartClient(config *config.TunConfig) {
	log.Debugf("vtun tcp client started on %v", config.LocalAddr)
	iface := tun.CreateTun(config)
	go tunToTcp(config, iface)
	for {
		conn, err := net.DialTimeout("tcp", config.ServerAddr, time.Duration(config.Timeout)*time.Second)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		cache.GetCache().Set("tcpconn", conn, 24*time.Hour)
		tcpToTun(config, conn, iface)
		cache.GetCache().Delete("tcpconn")

	}
}

// 发送
func tunToTcp(config *config.TunConfig, iface *water.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		if v, ok := cache.GetCache().Get("tcpconn"); ok {
			b := packet[:n]
			if config.Obfs {
				packet = cipher.XOR(packet)
			}
			tcpconn := v.(net.Conn)
			tcpconn.SetWriteDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
			_, err = tcpconn.Write(b)
			if err != nil {
				continue
			}
		}
	}
}

// 接收(网页内容)
func tcpToTun(config *config.TunConfig, tcpconn net.Conn, iface *water.Interface) {
	defer tcpconn.Close()
	packet := make([]byte, config.MTU)
	for {
		tcpconn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		n, err := tcpconn.Read(packet)
		if err != nil || err == io.EOF {
			break
		}
		b := packet[:n]
		if config.Obfs {
			b = cipher.XOR(b)
		}
		_, err = iface.Write(b)
		if err != nil {
			break
		}
	}
}
