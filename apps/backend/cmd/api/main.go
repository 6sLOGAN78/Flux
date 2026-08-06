// Package main provides the Echo v4 REST API server entrypoint for Flux backend.
package main

import (
	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/server"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load application configuration")
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize server")
	}
	defer func() {
		if srv.DBPool != nil {
			srv.DBPool.Close()
		}
	}()

	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("server terminated unexpectedly")
	}
}
