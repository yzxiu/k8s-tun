package ws

import (
	"net"
	"time"

	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/cache"
	"github.com/yzxiu/k8s-tun/pkg/common/cipher"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"github.com/yzxiu/k8s-tun/pkg/tun"
)

var firstInit = true

// Start websocket client
func StartClient(config *config.TunConfig) {
	log.Debugf("vtun websocket client started on %v", config.LocalAddr)
	iface := tun.CreateTun(config)
	go tunToWs(config, iface)
	for {
		conn := netutil.ConnectServer(config)
		if conn == nil {
			time.Sleep(3 * time.Second)
			continue
		} else {
			if firstInit {
				log.Infof("连接成功")
				firstInit = false
			}
		}
		cache.GetCache().Set("wsconn", conn, 24*time.Hour)
		wsToTun(config, conn, iface)
		cache.GetCache().Delete("wsconn")
	}
}

func wsToTun(config *config.TunConfig, wsconn net.Conn, iface *water.Interface) {
	defer wsconn.Close()
	for {
		wsconn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		packet, err := wsutil.ReadServerBinary(wsconn)
		if err != nil {
			break
		}
		if config.Obfs {
			packet = cipher.XOR(packet)
		}
		_, err = iface.Write(packet)
		if err != nil {
			break
		}
	}
}

func tunToWs(config *config.TunConfig, iface *water.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		if v, ok := cache.GetCache().Get("wsconn"); ok {
			b := packet[:n]
			if config.Obfs {
				packet = cipher.XOR(packet)
			}
			wsconn := v.(net.Conn)
			wsconn.SetWriteDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
			if err = wsutil.WriteClientBinary(wsconn, b); err != nil {
				continue
			}
		}
	}
}
