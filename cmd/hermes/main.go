package main

import (
	"fmt"
	"log"

	"github.com/thedeepak12/hermes-dns/internal/config"
)

func main() {
	fmt.Println("HermesDNS")

	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Server listening on: %s", cfg.Server.Listen)
	log.Printf("Upstream providers: %v", cfg.Upstream.Providers)

	log.Println("HermesDNS started successfully")
}
