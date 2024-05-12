package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/vlasashk/websocket-chat/config"
	"github.com/vlasashk/websocket-chat/internal/storage"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewStorageCfg()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err = storage.Run(ctx, cfg); err != nil {
		log.Fatal().Err(err).Send()
	}
}
