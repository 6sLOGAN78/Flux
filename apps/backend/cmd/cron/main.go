// Package main provides the background job queue worker for Flux backend.
package main

import (
	"log"

	"flux/apps/backend/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.Printf("starting flux background job worker processing tasks on redis %s...", cfg.GetRedisURL())
	// Asynq worker listening logic initialized here
}
