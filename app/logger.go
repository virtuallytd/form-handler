package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Initialize the logger
func init() {
	if _, err := os.Stat("/app/logs"); os.IsNotExist(err) {
		if err := os.Mkdir("/app/logs", os.ModePerm); err != nil {
			log.Fatalf("Could not create logs directory: %v", err)
		}
	}

	logFile, err := os.OpenFile("/app/logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.Out = logFile
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
}
