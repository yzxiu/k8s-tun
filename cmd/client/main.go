package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/yzxiu/k8s-tun/cmd/vtun/app"
	"github.com/yzxiu/k8s-tun/pkg/client"
)

var (
	infoUrl             string
	selectedEnv         string
	uuidFile            bool
	logLevel            int
	heartbeatPeriod     int
	heartbeatRetryTimes int
)

func main() {
	flag.IntVar(&logLevel, "v", 4, "log level, 4 is info, 5 is debug")
	flag.StringVar(&selectedEnv, "e", "", "env to client")
	flag.StringVar(&infoUrl, "s", "", "info server url")
	flag.BoolVar(&uuidFile, "u", false, "use uuid file")
	flag.IntVar(&heartbeatPeriod, "h", 60, "heartbeat period")
	flag.IntVar(&heartbeatRetryTimes, "t", 10, "heartbeat retry times")

	flag.Parse()
	log.SetLevel(log.Level(logLevel))
	// client
	c := client.GetClient(infoUrl)
	// uuid
	u := client.GetUUID(uuidFile)
	// select config
	env := client.SelectConfig(selectedEnv, c.Envs)
	// get server info
	sInfo := client.GetServerInfo(env, u)
	// get tun config
	tunConfig := client.GetTunConfig(sInfo)
	// heartbeat
	stopCh := client.HeartBeat(env, sInfo, u, heartbeatPeriod, heartbeatRetryTimes)
	// start tun
	app.StartTun(tunConfig, stopCh)
}
