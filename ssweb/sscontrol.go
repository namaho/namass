package main

import (
	"bytes"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type CmdDataAddPort struct {
	Port     int
	Password string
}

type CmdDataDelPort struct {
	Port int
}

func SSCmdAddPort(ctx context.Context, port int, password string) error {
	return SSCmdExec(ctx, "add_port", CmdDataAddPort{Port: port, Password: password})
}

func SSCmdDelPort(ctx context.Context, port int) error {
	return SSCmdExec(ctx, "del_port", CmdDataDelPort{Port: port})
}

func SSCmdExec(ctx context.Context, cmd string, data interface{}) error {
	config := ctx.Value("config").(Config)
	var err error

	ips, err := GetServerIPList(ctx)
	if err != nil {
		log.Error("fail to get server list while adding port for user")
		return err
	}

	for _, ip := range ips {
		go func(ip string, config Config) {
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(data)
			resp, err := http.Post("http://"+ip+":"+config.SSAgentHTTPPort+"/ssagent/"+cmd, "application/json; charset=utf-8", b)
			if err != nil {
				log.Error(err)
				return
			}

			if resp.Body == nil {
				log.Error("can not connect server " + ip)
				return
			}

			bs, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
			}

			log.Debug(string(bs))
		}(ip, config)
	}

	return err
}
