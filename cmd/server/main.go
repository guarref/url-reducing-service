package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/guarref/url-reducing-service/config"
	"github.com/guarref/url-reducing-service/internal/app"
)

func main() {

	storageFlag := flag.String("storage", "", "choose the storage flag: memory or postgres")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config")
	}

	if *storageFlag != "" {
		cfg.Service.Storage = *storageFlag
		if err := cfg.Validate(); err != nil {
			log.Fatal().Err(err).Msg("invalid config from flag")
		}
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating new application")
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
