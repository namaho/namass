package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"path"
)

func SetupLog(config Config) {
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
