package info

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"net/http"
	"time"
)

var info config.ServerInfo
var tunConfig config.TunConfig
var kubeClient *kubernetes.Clientset
var dyClient dynamic.Interface

var ipManager IpManage

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&info.Kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&info.MasterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&info.SvcCIDR, "svccidr", "", "")
	flag.StringVar(&info.PodCIDR, "podcidr", "", "")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(info.MasterURL, info.Kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}

	dyClient, err = dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}
}

func Server() *http.Server {
	// info update from env
	tunConfig.UpdateFromEnv()
	tunConfig.Validate()

	// init ip pool
	ipManager = NewIpManage(tunConfig.CIDR, kubeClient)

	// get service-cluster-ip-range
	svccidr := GetSvcCIDR(kubeClient)
	if len(svccidr) == 0 {
		log.Fatalf("Failed to get the [service-cluster-ip-range] parameter from the cluster, please configure the [-svccidr] parameters manually")
	} else {
		log.Debugf("service-cidr: %s", svccidr)
		info.SvcCIDR = svccidr
	}

	// get cluster-cidr
	podcidrs := GetPodCIDR(dyClient, kubeClient)
	if len(podcidrs) == 0 {
		log.Fatalf("Failed to get the [cluster-cidr] parameter from the cluster, please configure the [-podcidr] parameters manually")
	} else {
		log.Debugf("pod-cidr: %s", podcidrs)
		info.PodCIDR = podcidrs
	}

	// info server
	engine := gin.Default()
	server := &http.Server{
		Addr:    ":3002",
		Handler: engine,
	}
	engine.GET("info", getInfoHandler)
	engine.GET("validate", validateHandler)
	engine.GET("heartbeat", heartbeatHandler)
	return server
}

func validateHandler(c *gin.Context) {
	c.JSON(http.StatusOK, ipManager.Validate())
}

func getInfoHandler(c *gin.Context) {
	// tun server host/port
	tunServer, err := GetTunServer(kubeClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	// client ip
	uuid := c.Query("uuid")
	if len(uuid) < 1 {
		c.JSON(http.StatusInternalServerError, "Please check the [uuid] parameter.")
		return
	}
	clientCIDR := ipManager.GetIP(uuid)
	if len(clientCIDR) < 1 {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	rInfo := config.ServerInfo{
		Kubeconfig: "",
		MasterURL:  "",
		SvcCIDR:    info.SvcCIDR,
		PodCIDR:    info.PodCIDR,
		TunServer:  tunServer.Host + ":" + tunServer.Port,
		//TunServerPort: tunServer.Port,
		ClientCIDR:    clientCIDR,
		Key:           tunConfig.Key,
		Protocol:      tunConfig.Protocol,
		WebSocketPath: tunConfig.WebSocketPath,
		Obfs:          tunConfig.Obfs,
	}
	c.JSON(http.StatusOK, rInfo)
}

func heartbeatHandler(c *gin.Context) {
	uuid := c.Query("uuid")
	if len(uuid) < 1 {
		c.JSON(http.StatusInternalServerError, "Please check the [uuid] parameter.")
		return
	}
	ip := c.Query("ip")
	if len(ip) < 1 {
		c.JSON(http.StatusInternalServerError, "Please check the [ip] parameter.")
		return
	}
	c.JSON(http.StatusOK, ipManager.HeartBeat(uuid, ip))
}
