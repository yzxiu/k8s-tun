package info

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

type IP struct {
	Ip  string `json:"ip"`
	Exp *time.Time
}

type IpManage interface {
	GetIP(string) string
	Validate() map[string]interface{}
	HeartBeat(uuid string, ip string) *config.HeartbeatResult
}

type manager struct {
	mu         sync.Mutex
	to         string
	expDur     time.Duration
	vpnNetWork *net.IPNet
	clientSet  *kubernetes.Clientset

	ips   []string
	ipSet mapset.Set[string]
	used  map[string]IP // uuid:ip

	lastActive map[string]IP
}

func NewIpManage(cidr string, clientSet *kubernetes.Clientset) IpManage {
	// Vpn Network
	vpnNetwork := &net.IPNet{}
	if _, nw, err := net.ParseCIDR(cidr); err != nil {
		log.WithError(err).Fatalf("Parse cidr error: %s", cidr)
	} else {
		vpnNetwork = nw
	}

	// IP timeout
	var IpTimeOut = "15m"
	if to := os.Getenv("K8S_TUN_IP_TIMEOUT"); len(to) > 0 {
		IpTimeOut = to
	}
	log.Infof("vtun ip timeout: %s", IpTimeOut)
	expDur, err := time.ParseDuration(IpTimeOut)
	if err != nil {
		log.Fatalf("timeout paramter error: %s", IpTimeOut)
	}

	// Init ips
	ips, _ := netutil.CidrIps(cidr)
	// Remove the first 10 ips, begin with *.*.*.11/24
	ips = ips[10:]
	// Remove the lost 10 ips, end with *.*.*.249/24
	ips = ips[:len(ips)-5]
	ipSet := mapset.NewSet[string]()
	for _, ip := range ips {
		ipSet.Add(ip)
	}

	// Load data from [ConfigMap]
	used := map[string]IP{}
	datas := getConfigMap(clientSet).Data
	if len(datas) > 0 {
		for uuid, ip := range datas {
			if ipSet.Contains(ip) {
				ipSet.Remove(ip)
				used[uuid] = IP{
					Ip:  ip,
					Exp: getExpireTime(expDur),
				}
			} else {
				// There is no need to remove and update the [ConfigMap] here,
				// Just wait for the following syncData update
			}
		}
	}
	m := &manager{
		mu:         sync.Mutex{},
		to:         IpTimeOut,
		expDur:     expDur,
		vpnNetWork: vpnNetwork,
		clientSet:  clientSet,
		ips:        ips,
		ipSet:      ipSet,
		used:       used,
		lastActive: map[string]IP{},
	}
	go m.checkExpire()
	return m
}

func (m *manager) Validate() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	ipsMap := map[string]string{}
	for _, ip := range m.ips {
		ipsMap[ip] = ""
	}

	var conflictIPs []string
	for i := range m.ipSet.Iterator().C {
		ipsMap[i] = "0"

		// check conflictIPs
		if ip, _, err := net.ParseCIDR(i); err != nil {
			conflictIPs = append(conflictIPs, i)
		} else {
			if !m.vpnNetWork.Contains(ip) {
				conflictIPs = append(conflictIPs, i)
			}
		}
	}
	var duplicateIPs []string
	duplicateIPsMap := map[string]int{}
	for _, ip := range m.used {
		ipsMap[ip.Ip] = "1"

		// check conflictIPs
		if ipParse, _, err := net.ParseCIDR(ip.Ip); err != nil {
			conflictIPs = append(conflictIPs, ip.Ip)
		} else {
			if !m.vpnNetWork.Contains(ipParse) {
				conflictIPs = append(conflictIPs, ip.Ip)
			}
		}

		// check duplicateIPs
		if cou, ok := duplicateIPsMap[ip.Ip]; ok {
			duplicateIPsMap[ip.Ip] = cou + 1
		} else {
			duplicateIPsMap[ip.Ip] = 1
		}
	}
	for ip, count := range duplicateIPsMap {
		if count > 1 {
			duplicateIPs = append(duplicateIPs, ip)
		}
	}
	validate := true
	for _, value := range ipsMap {
		if len(value) == 0 {
			validate = false
		}
	}
	allCount := len(m.ips)
	used := len(m.used)
	unused := m.ipSet.Cardinality()
	return map[string]interface{}{
		"activeIPs":    m.used,
		"serverTime:":  time.Now(),
		"ipTimeout":    m.to,
		"conflictIPs":  conflictIPs,
		"duplicateIPs": duplicateIPs,
		"allCount":     allCount,
		"usedCount":    used,
		"unusedCount":  unused,
		"validate":     validate && (used+unused) == allCount && len(conflictIPs) == 0 && len(duplicateIPs) == 0,
	}
}

func (m *manager) GetIP(uuid string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Determine if [uuid] has assigned an ip,
	// if so, update expiration time and return
	if ip, ok := m.used[uuid]; ok {
		m.used[uuid] = IP{
			Ip:  ip.Ip,
			Exp: m.getExpireTime(),
		}
		return ip.Ip
	}

	// Pop an ip from m.ipSet,
	// and add to m.used
	ip, succ := m.ipSet.Pop()
	if succ {
		log.Infof("Assigned an [ip:%s] to [uuid:%s]", ip, uuid)
		m.used[uuid] = IP{
			Ip:  ip,
			Exp: m.getExpireTime(),
		}
		return ip
	}

	return ""
}

// HeartBeat If the heartbeat fails,
// the client should delete the uuid(if uuid file exist) and exit
func (m *manager) HeartBeat(uuid string, ip string) *config.HeartbeatResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	if contain, _ := m.contains(ip); !contain {
		delete(m.used, uuid)
		log.Errorf("heartbeat ip[%s] error", ip)
		return heartbeatResult(config.HeartbeatIPNotContain)
	}

	if m.heartbeatIPConflict(uuid, ip) {
		log.Errorf("heartbeat ip[%s] conflict", ip)
		return heartbeatResult(config.HeartbeatIPConflict)
	}

	if exist, oldIP := m.heartbeatUUIDConflict(uuid, ip); exist {
		log.Errorf("heartbeat [uuid: %s ip:%s] conflict, oldIP:%s", uuid, ip, oldIP)
		return heartbeatResult(config.HeartbeatUUIDConflict)
	}

	if m.heartbeatNotFound(uuid, ip) {
		log.Warnf("heartbeat [uuid: %s ip:%s] not found, add it to used map", uuid, ip)
		return heartbeatResult(config.HeartbeatNotFound)
	}

	if m.heartbeatCheckedAndUpdateExpireTime(uuid, ip) {
		log.Debugf("heartbeat [uuid: %s ip:%s] check success", uuid, ip)
		return heartbeatResult(config.HeartbeatSuccess)
	}

	log.Errorf("heartbeat [uuid: %s ip:%s] error, unknown", uuid, ip)
	return heartbeatResult(config.HeartbeatUnknownError)
}

func heartbeatResult(result string) *config.HeartbeatResult {
	return &config.HeartbeatResult{
		Result: result,
	}
}

func (m *manager) contains(ip string) (bool, error) {
	if ip1, _, err := net.ParseCIDR(ip); err != nil {
		log.WithError(err).Warnf("ip[%s] error", ip)
		return false, err
	} else {
		if m.vpnNetWork.Contains(ip1) {
			return true, nil
		}
		return false, nil
	}
}

func (m *manager) heartbeatIPConflict(uuid string, ip string) bool {
	for u, i := range m.used {
		if i.Ip == ip {
			if uuid != u {
				return true
			}
		}
	}
	return false
}

func (m *manager) heartbeatUUIDConflict(uuid string, ip string) (bool, string) {
	for u, oldIP := range m.used {
		if u == uuid {
			if oldIP.Ip != ip {
				return true, oldIP.Ip
			}
		}
	}
	return false, ""
}

func (m *manager) heartbeatNotFound(uuid string, ip string) bool {
	for u, i := range m.used {
		if u == uuid || i.Ip == ip {
			return false
		}
	}
	// 心跳发送过来的uuid和ip都未使用，则添加到used
	if m.ipSet.Contains(ip) {
		m.ipSet.Remove(ip)
		m.used[uuid] = IP{
			Ip:  ip,
			Exp: m.getExpireTime(),
		}
		return true
	}
	return false
}

func (m *manager) heartbeatCheckedAndUpdateExpireTime(uuid string, ip string) bool {
	for u, i := range m.used {
		if u == uuid {
			if i.Ip == ip {
				m.used[u] = IP{
					Ip:  i.Ip,
					Exp: m.getExpireTime(),
				}
			}
			return true
		}
	}
	return false
}

func (m *manager) checkExpire() {
	for true {
		time.Sleep(time.Second * 5)
		activeMap, updated := m.checkExpireIP()
		// update used map to configmap
		if updated {
			m.syncData(activeMap)
		}
	}
}

func (m *manager) checkExpireIP() (map[string]IP, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	activeMap := map[string]IP{}
	for uuid, ip := range m.used {
		// Expired
		if time.Now().After(*ip.Exp) {
			log.Infof("[ip:%s uuid:%s] is timeout", ip.Ip, uuid)
			delete(m.used, uuid)
			// put back
			m.ipSet.Add(ip.Ip)
		} else {
			activeMap[uuid] = ip
		}
	}
	// activeMap != 0 时才不需要更新,
	// 否则会出现configmap中的内容没有被刷新掉的情况
	// TODO 当 len(activeMap) == 0 时，每次 checkExpireIP 都会刷新configmap，导致刷新频繁
	if reflect.DeepEqual(m.lastActive, activeMap) && len(activeMap) != 0 {
		return nil, false
	}
	m.lastActive = activeMap
	return activeMap, true
}

// syncData
func (m *manager) syncData(activeIps map[string]IP) {
	data := map[string]string{}
	for uuid, ip := range activeIps {
		data[uuid] = ip.Ip
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := m.clientSet.CoreV1().ConfigMaps("k8s-tun").Update(ctx, &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:      "k8s-tun-active-ips",
			Namespace: "k8s-tun",
		},
		Data: data,
	}, v1.UpdateOptions{})
	if err != nil {
		log.WithError(err).Warnf("Update active ips to ConfigMap error: %v", activeIps)
	}
}

func (m *manager) getExpireTime() *time.Time {
	return getExpireTime(m.expDur)
}

func getExpireTime(dur time.Duration) *time.Time {
	expTime := time.Now().Add(dur)
	return &expTime
}

func getConfigMap(clientSet *kubernetes.Clientset) *corev1.ConfigMap {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if configMap, err := clientSet.CoreV1().ConfigMaps("k8s-tun").Get(ctx, "k8s-tun-active-ips", v1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			// create
			cf, err := clientSet.CoreV1().ConfigMaps("k8s-tun").Create(ctx, &corev1.ConfigMap{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name:      "k8s-tun-active-ips",
					Namespace: "k8s-tun",
				},
			}, v1.CreateOptions{})
			if err != nil {
				log.WithError(err).Fatalf("%v", err)
			}
			return cf
		} else {
			log.WithError(err).Fatalf("%v", err)
			return nil
		}
	} else {
		return configMap
	}
}
