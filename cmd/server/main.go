package main

import (
	"log"
	"os"

	"green-api/internal/app"
)

func main() {
	configPath := os.Getenv("APP_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	server, err := app.New(configPath)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
