package sproxy

import (
	"fmt"
)

func Test() {
	pconf := new(ProxyConfig)
	conf, _:= pconf.ReadConf("sproxy.conf")
	if conf == nil {
		fmt.Println("no conf")
		return
	}

	log := GetLogger(conf.log, conf.level)

	server, err := InitServer(conf, log)
	if err != nil {
		log.Error("init server failed")
		return
	}

	server.CoreRun()
}

