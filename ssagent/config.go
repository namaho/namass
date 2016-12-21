package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

type Config struct {
	Listen           string `json:"listen"`
	LogLevel         string `json:"loglevel"`
	LogFile          string `json:"logfile"`
	SSAgentUnixSock  string `json:"ssagent_unix_sock"`
	SSDaemonUnixSock string `json:"ssdaemon_unix_sock"`
	SSWebHTTPAddress string `json:"ssweb_http_address"`
	ServerToken      string `json:"server_token"`
	ReportInterface  string `json:"report_interface"`
	Area             string `json:"area"`
}

func SetupLog() {
	switch config.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	logFilePath := config.LogFile

	if config.LogFile == "stdout" {
		log.SetOutput(os.Stdout)
		return
	}

	if config.LogFile == "stderr" {
		log.SetOutput(os.Stderr)
		return
	}

	if !path.IsAbs(config.LogFile) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		logFilePath = path.Join(wd, logFilePath)
	}

	err := os.MkdirAll(path.Dir(logFilePath), 0755)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
}

func ReadConfigFile(configFile string) ([]byte, error) {
	var configStr []byte
	configFiles := [3]string{configFile, "cfg.json", "/etc/ssagent.json"}
	for _, file := range configFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}
		configStr, err := ioutil.ReadFile(file)
		return configStr, err
	}
	return configStr, errors.New("Could not find config file")
}
