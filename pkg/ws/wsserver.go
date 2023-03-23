package ws

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/yzxiu/k8s-tun/pkg/common/cache"
	"github.com/yzxiu/k8s-tun/pkg/common/cipher"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/counter"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"github.com/yzxiu/k8s-tun/pkg/tun"
)

// Start websocket server
func StartServer(config *config.TunConfig) {
	iface := tun.CreateTun(config)
	// server -> client
	go toClient(config, iface)
	// client -> server
	http.HandleFunc(config.WebSocketPath, func(w http.ResponseWriter, r *http.Request) {
		if !checkPermission(w, r, config) {
			return
		}
		wsconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Debugf("[server] failed to upgrade http %v", err)
			return
		}
		toServer(config, wsconn, iface)
	})

	log.Debugf("vtun websocket server started on %v", config.LocalAddr)
	http.ListenAndServe(config.LocalAddr, nil)
}
func checkPermission(w http.ResponseWriter, req *http.Request, config *config.TunConfig) bool {
	key := req.Header.Get("key")
	if key != config.Key {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("No permission"))
		return false
	}
	return true
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
				wsutil.WriteServerBinary(v.(net.Conn), b)
			}
		}
	}
}

func toServer(config *config.TunConfig, wsconn net.Conn, iface *water.Interface) {
	defer wsconn.Close()
	for {
		wsconn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		b, err := wsutil.ReadClientBinary(wsconn)
		if err != nil || err == io.EOF {
			break
		}
		if config.Obfs {
			b = cipher.XOR(b)
		}
		if key := netutil.GetSourceKey(b); key != "" {
			cache.GetCache().Set(key, wsconn, 10*time.Minute)
			counter.IncrReadByte(len(b))
			iface.Write(b)
		}
	}
}
