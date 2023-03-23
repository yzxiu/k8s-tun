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
	"github.com/yzxiu/k8s-tun/pkg/common/counter"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"github.com/yzxiu/k8s-tun/pkg/tun"
)

// Start tcp server
func StartServer(config *config.TunConfig) {
	log.Debugf("vtun tcp server started on %v", config.LocalAddr)
	iface := tun.CreateTun(config)
	// server -> client
	go toClient(config, iface)
	ln, err := net.Listen("tcp", config.LocalAddr)
	if err != nil {
		log.Error(err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		log.Debugf("连接请求: loAddr: %s  roAddr: %s", conn.LocalAddr().String(), conn.RemoteAddr().String())
		// client -> server
		go toServer(config, conn, iface)
	}

}

func toClient(config *config.TunConfig, iface *water.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || err == io.EOF || n == 0 {
			continue
		}
		b := packet[:n]
		if key := netutil.GetDestinationKey(b); key != "" {
			if v, ok := cache.GetCache().Get(key); ok {
				if config.Obfs {
					b = cipher.XOR(b)
				}
				counter.IncrWriteByte(n)
				v.(net.Conn).Write(b)
			}
		}
	}
}

func toServer(config *config.TunConfig, tcpconn net.Conn, iface *water.Interface) {
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
		if key := netutil.GetSourceKey(b); key != "" {
			cache.GetCache().Set(key, tcpconn, 10*time.Minute)
			counter.IncrReadByte(len(b))
			iface.Write(b)
		}
	}
}
