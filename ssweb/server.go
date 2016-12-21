package main

import (
	"context"
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

type Server struct {
	IP   string `json:"ip"`
	Area int    `json:"area"`
}

// Handle /ssweb/server/discover request. The ssagent will report host information
// to this API after startup, ssweb response with a Shadowsocks port list
// including ("port":"password") pairs which would be applied to the target
// Shadowsocks daemon.
func ServerDiscover(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var server Server
	err := json.NewDecoder(r.Body).Decode(&server)
	if err != nil {
		log.Error(err)
	}

	ip := server.IP
	area := server.Area

	_, err = GetServerIdByIP(ctx, ip)
	if err != nil {
		if err == sql.ErrNoRows {
			err = SaveServer(ctx, ip, area)
			if err != nil {
				log.Error(err)
			} else {
				log.Info("save new server " + ip)
			}
		} else {
			log.Error(err)
		}
	} else {
		log.Warn("server " + ip + " already exists")
	}

	portList, err := GetSSPortList(ctx)
	configStr, err := json.Marshal(portList)
	if err != nil {
		log.Error(err)
	}

	log.Info("send ssport config to ssagent")
	w.Header().Set("Content-Type", "application/json")
	w.Write(configStr)
}
