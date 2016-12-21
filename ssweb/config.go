package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type DB struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type SMTP struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Port     string `json:"port"`
}

type Config struct {
	Listen           string `json:"listen"`
	LogLevel         string `json:"loglevel"`
	LogFile          string `json:"logfile"`
	SSPasswordPrefix string `json:"ss_password_prefix"`
	SSWebHTTPAddress string `json:"ssweb_http_address"`
	SSAgentHTTPPort  string `json:"ssagent_http_port"`
	SSServerDomain   string `json:"ss_server_domain"`
	ServerToken      string `json:"server_token"`
	DB               DB     `json:"db"`
	Smtp             SMTP   `json:"smtp"`
}

func ReadConfigFile(configFile string) (Config, error) {
	configFiles := [3]string{configFile, "cfg.json", "/etc/ssweb.json"}
	for _, file := range configFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}
		configFile, err := ioutil.ReadFile(file)

		var config Config
		err = json.Unmarshal(configFile, &config)
		if err != nil {
			return Config{}, err
		}
		return config, nil
	}
	return Config{}, errors.New("no config files found")
}
