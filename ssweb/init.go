package main

import (
	"context"
	"flag"
	log "github.com/Sirupsen/logrus"
)

func Init() context.Context {
	var configFile string
	var ctx context.Context

	flag.StringVar(&configFile, "c", "cfg.json", "configuration file")
	flag.Parse()

	config, err := ReadConfigFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	SetupLog(config)

	db, err := NewDB(config)
	if err != nil {
		log.Panic(err)
	}

	ctx = context.WithValue(context.Background(), "db", db)
	ctx = context.WithValue(ctx, "config", config)

	return ctx
}
