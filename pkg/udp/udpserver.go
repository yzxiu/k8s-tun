package udp

import (
	"net"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/cipher"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"github.com/yzxiu/k8s-tun/pkg/tun"
)

// Start udp server
func StartServer(config *config.TunConfig) {
	log.Debugf("vtun udp server started on %v", config.LocalAddr)
	iface := tun.CreateTun(config)
	localAddr, err := net.ResolveUDPAddr("udp", config.LocalAddr)
	if err != nil {
		log.Fatalln("failed to get udp socket:", err)
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("failed to listen on udp socket:", err)
	}
	defer conn.Close()
	// server -> client
	reply := &Reply{localConn: conn, connCache: cache.New(30*time.Minute, 10*time.Minute)}
	go reply.toClient(config, iface, conn)
	// client -> server
	packet := make([]byte, config.MTU)
	for {
		n, cliAddr, err := conn.ReadFromUDP(packet)
		if err != nil || n == 0 {
			continue
		}
		var b []byte
		if config.Obfs {
			b = cipher.XOR(packet[:n])
		} else {
			b = packet[:n]
		}
		if key := netutil.GetSourceKey(b); key != "" {
			iface.Write(b)
			reply.connCache.Set(key, cliAddr, cache.DefaultExpiration)
		}
	}
}

type Reply struct {
	localConn *net.UDPConn
	connCache *cache.Cache
}

func (r *Reply) toClient(config *config.TunConfig, iface *water.Interface, conn *net.UDPConn) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		b := packet[:n]
		if key := netutil.GetDestinationKey(b); key != "" {
			if v, ok := r.connCache.Get(key); ok {
				if config.Obfs {
					b = cipher.XOR(b)
				}
				r.localConn.WriteToUDP(b, v.(*net.UDPAddr))
			}
		}
	}
}
