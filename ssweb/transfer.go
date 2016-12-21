package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
)

// Handle /ssweb/transfer/report request, collect Shadowsocks ports transfer
// statistics from ssagent.
func ReportTransfer(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	b := new(bytes.Buffer)
	b.ReadFrom(r.Body)
	stat := map[string]int{}
	log.Debug("transfer stat: " + b.String())
	err := json.Unmarshal(b.Bytes(), &stat)
	if err != nil {
		log.Error(err)
		fmt.Fprintf(w, "fail")
	}

	for port, transfer := range stat {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, "fail")
		}

		err = UpdateTransfer(ctx, portInt, transfer)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, "fail")
		}
	}

	fmt.Fprintf(w, "ok")
}
