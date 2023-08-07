package client

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yzxiu/k8s-tun/cmd/client/envs"
	"github.com/yzxiu/k8s-tun/pkg/common/config"
	"github.com/yzxiu/k8s-tun/pkg/common/netutil"
	"github.com/yzxiu/k8s-tun/pkg/common/signal"
	"gopkg.in/yaml.v2"
)

type Client struct {
	Envs []*Env `yaml:"envs"`
}

type Env struct {
	Name    string `yaml:"name"`
	InfoUrl string `yaml:"infoUrl"`
}

func GetClient(infoUrl string) *Client {
	c := new(Client)
	if len(infoUrl) > 0 {
		c.Envs = []*Env{
			{
				Name:    "server",
				InfoUrl: infoUrl,
			},
		}
		return c
	}
	c = loadYamlFile()
	if c == nil {
		c = &Client{}
		err := yaml.UnmarshalStrict([]byte(envs.InitEnvs), c)
		if err != nil || c == nil || c.Envs == nil || len(c.Envs) == 0 {
			log.Fatalf("load client config err")
		}
	}
	return c
}

func loadYamlFile() *Client {
	fs, err := filepath.Abs("client.yaml")
	if err != nil {
		log.Fatalf("read client config err")
	}
	yamlFile, err := ioutil.ReadFile(fs)
	if err != nil {
		return nil
	}
	if len(yamlFile) == 0 {
		return nil
	}
	c := &Client{}
	err = yaml.UnmarshalStrict(yamlFile, c)
	if err != nil || c == nil || c.Envs == nil || len(c.Envs) == 0 {
		return nil
	}
	return c
}

func SelectConfig(selectedEnv string, envs []*Env) *Env {
	if len(envs) == 1 {
		return envs[0]
	}
	envName := ""
	var es []string
	for _, n := range envs {
		es = append(es, n.Name)
	}
	if len(selectedEnv) > 0 && exist(es, selectedEnv) {
		envName = selectedEnv
	} else {
		prompt := &survey.Select{
			Message: "请选择需要连接的环境:",
			Options: es,
		}
		err := survey.AskOne(prompt, &envName)
		if err != nil {
			log.Fatalf("select config error")
		}
	}
	for _, e := range envs {
		if e.Name == envName {
			return e
		}
	}
	log.Fatalf("select config error")
	return nil
}

func exist(ss []string, s string) bool {
	for _, sub := range ss {
		if sub == s {
			return true
		}
	}
	return false
}

func GetServerInfo(env *Env, uuid string) *config.ServerInfo {
	info := &config.ServerInfo{}
	resp, err := netutil.New().Query("uuid", uuid).Get("http://" + env.InfoUrl + "/" + "info")
	if err != nil {
		log.WithError(err).Fatalf("get server info err")
	}
	err = resp.AsJSON(info)
	if len(info.TunServer) == 0 {
		log.Fatalf("tun server error")
	}
	if err != nil {
		log.WithError(err).Fatalf("info server data error")
	}
	return info
}

func HeartBeat(env *Env, info *config.ServerInfo, uuid string, period int, times int) <-chan struct{} {
	if period < 5 {
		period = 5
	}
	if times <= 3 {
		times = 3
	}
	log.Debugf("heartbeat period: %ds", period)
	log.Debugf("heartbeat retry times: %d", times)

	heartbeatStopCh := make(chan struct{})
	go func() {
		<-signal.SetupSignalHandler()
		close(heartbeatStopCh)
	}()
	go beat(env, info, uuid, heartbeatStopCh, period, times)
	return heartbeatStopCh
}

func beat(env *Env, info *config.ServerInfo, uuid string, heartbeatStopCh chan struct{}, period int, times int) {
	heartbeatErrorTimes := 0
	for {
		time.Sleep(time.Second * time.Duration(period))
		result := &config.HeartbeatResult{}
		resp, err := netutil.New().Query("uuid", uuid).Query("ip", info.ClientCIDR).Get("http://" + env.InfoUrl + "/" + "heartbeat")
		if err != nil || resp == nil {
			CloseClient(&heartbeatErrorTimes, times, heartbeatStopCh)
			continue
		}
		err = resp.AsJSON(result)
		if err != nil {
			log.WithError(err).Errorf("Heartbeat Failed: %v", err)
			CloseClient(&heartbeatErrorTimes, times, heartbeatStopCh)
			continue
		}
		if result.Result == config.HeartbeatSuccess || result.Result == config.HeartbeatNotFound {
			log.Debugf("Heartbeat Success: %s", result.Result)
		} else {
			log.Errorf("Heartbeat conflict: %s, please restart client", result.Result)
			close(heartbeatStopCh)
			return
		}
		heartbeatErrorTimes = 0
	}
}

func CloseClient(count *int, retryTimes int, stop chan struct{}) {
	*count = *count + 1
	log.Errorf("Heartbeat failed : %d", *count)
	if *count >= retryTimes {
		close(stop)
	}
}

func GetTunConfig(sInfo *config.ServerInfo) *config.TunConfig {
	//-l=:3000 -s=192.168.242.81:30010 -c=10.99.99.10/24 -k=123456 -p ws
	tunConfig := &config.TunConfig{}
	tunConfig.CIDR = sInfo.ClientCIDR
	tunConfig.DstCIDR = sInfo.SvcCIDR + "," + sInfo.PodCIDR
	tunConfig.MTU = 1500
	tunConfig.LocalAddr = ":3000"
	tunConfig.ServerAddr = sInfo.TunServer
	tunConfig.Key = sInfo.Key
	tunConfig.Protocol = sInfo.Protocol
	tunConfig.WebSocketPath = sInfo.WebSocketPath
	tunConfig.ServerMode = false
	tunConfig.GlobalMode = false
	tunConfig.Obfs = sInfo.Obfs
	tunConfig.Timeout = 30
	return tunConfig
}
